package rdpg

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/wayneeseguin/rdpg-agent/log"
)

// TODO: RDPG Struct => RDPG Struct, allowing for multiple instances of RDPG
func NewRDPG(uri string) *RDPG {
	if uri == "" || uri[0:13] != "postgresql://" {
		log.Error(fmt.Sprintf("rdpg.NewRDPG() uri malformed ! %s", uri))
		return nil
	}
	return &RDPG{URI: uri}
}

func (r *RDPG) connect() (db *sqlx.DB, err error) {
	db, err = sqlx.Connect("postgres", r.URI)
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.Node#Connect() %s ! %s", r.URI, err))
	}
	return db, err
}

func (r *RDPG) Nodes() (nodes []Node) {
	db, err := r.connect()
	if err != nil {
		log.Error(fmt.Sprintf("RDPG#Nodes() ! %s", err))
	}

	// TODO: Populate list of rdpg nodes for given URL,
	//`SELECT node_local_dsn FROM bdr.bdr_nodes INTO rdpg.nodes (node_local_dsn);`

	type dsn struct {
		DSN string `db:"node_local_dsn"`
	}

	dsns := []dsn{}
	err = db.Select(&dsns, SQL["bdr_nodes_dsn"])
	if err != nil {
		log.Error(fmt.Sprintf("RDPG#Nodes() %s ! %s", SQL["bdr_nodes"], err))
	}

	for _, t := range dsns {
		node := Node{}
		s := strings.Split(t.DSN, " ")
		node.LocalDSN = t.DSN
		node.Host = strings.Split(s[0], "=")[1]
		node.Port = strings.Split(s[1], "=")[1]
		node.User = strings.Split(s[2], "=")[1]
		node.Database = `postgres` // strings.Split(s[3], "=")[1]
		nodes = append(nodes, node)
	}
	// TODO: Get this information into the database and then out of the rdpg.nodes
	//rows, err := db.Query("SELECT host,port,user,'postgres' FROM rdpg.nodes;")
	//if err != nil {
	//	log.Error(fmt.Sprintf("Nodes() %s", err))
	//} else {
	//	sqlx.StructScan(rows, nodes)
	//}
	db.Close()
	return nodes
}

func (r *RDPG) CreateUser(username, password string) (err error) {
	for _, node := range r.Nodes() {
		node.Database = `postgres`
		db, err := node.Connect()
		if err != nil {
			log.Error(fmt.Sprintf(`RDPG#CreateUser(%s) %s ! %s`, username, node.Host, err))
			return err
		}

		// TODO: Check if user exists first
		sq := fmt.Sprintf(`CREATE USER %s;`, username)
		log.Trace(fmt.Sprintf(`RDPG#CreateUser(%s) %s > %s`, username, node.Host, sq))
		_, err = db.Exec(sq)
		if err != nil {
			log.Error(fmt.Sprintf("RDPG#CreateUser(%s) %s ! %s", username, node.Host, err))
			db.Close()
			return err
		}

		sq = fmt.Sprintf(`ALTER USER %s ENCRYPTED PASSWORD '%s'`, username, password)
		log.Trace(fmt.Sprintf(`RDPG#CreateUser(%s) %s > %s`, username, node.Host, sq))
		_, err = db.Exec(sq)
		if err != nil {
			log.Error(fmt.Sprintf("RDPG#CreateUser(%s) %s ! %s", username, node.Host, err))
		}
		db.Close()
	}
	return nil
}

func (r *RDPG) CreateDatabase(dbname, owner string) (err error) {
	for _, node := range r.Nodes() {
		err = node.CreateDatabase(dbname, owner)
		if err != nil {
			break
		}

		node.Database = dbname
		err = node.CreateExtensions([]string{"btree_gist", "bdr"})
		if err != nil {
			break
		}

	}
	if err != nil {
		// Cleanup in BDR currently requires droping the database and trying again...
		r.DropDatabase(dbname)
	}
	return err
}

func (r *RDPG) CreateReplicationGroup(dbname string) (err error) {
	nodes := r.Nodes()
	// TODO: Drop Database on all nodes if err != nil for any operation below
	for index, node := range nodes {
		node.Database = dbname
		db, err := node.Connect()
		if err != nil {
			break
		}
		sq := ""
		name := fmt.Sprintf("%s", node.Host)
		if index == 0 {
			sq = fmt.Sprintf(`SELECT bdr.bdr_group_create(
				local_node_name := '%s',
				node_external_dsn := 'host=%s port=%s user=%s dbname=%s'
			); `, name, node.Host, node.Port, node.User, dbname)
		} else {
			sq = fmt.Sprintf(`SELECT bdr.bdr_group_join(
				local_node_name := '%s',
				node_external_dsn := 'host=%s port=%s user=%s dbname=%s',
				join_using_dsn := 'host=%s port=%s user=%s dbname=%s'
			); `,
				name, node.Host, node.Port, node.User, node.Database,
				nodes[0].Host, nodes[0].Port, nodes[0].User, dbname,
			)
		}
		log.Trace(fmt.Sprintf(`RDPG#CreateReplicationGroup(%s) %s > %s`, dbname, node.Host, sq))
		_, err = db.Exec(sq)
		if err == nil {
			sq = `SELECT bdr.bdr_node_join_wait_for_ready();`
			log.Trace(fmt.Sprintf(`RDPG#CreateReplicationGroup(%s) %s ! %s`, dbname, node.Host, sq))
			_, err = db.Exec(sq)
		}
		db.Close()
	}
	if err != nil {
		// Cleanup in BDR currently requires droping the database and trying again...
		log.Error(fmt.Sprintf("CreateReplicationGroup(%s) ! %s", dbname, err))
	}
	return err
}

func (r *RDPG) DisableDatabase(dbname string) (err error) {
	for _, node := range r.Nodes() {
		node.Database = "postgres"
		db, err := node.Connect()
		if err != nil {
			log.Error(fmt.Sprintf("CreateReplicationGroup(%s) %s ! %s", dbname, node.Host, err))
			return err
		}

		sq := fmt.Sprintf(`UPDATE pg_database SET datallowconn = 'false' WHERE datname = '%s';`, dbname)
		log.Trace(fmt.Sprintf(`RDPG#DisableDatabase(%s) DISALLOW %s > %s`, dbname, node.Host, sq))
		_, err = db.Exec(sq)
		if err != nil {
			log.Error(fmt.Sprintf("RDPG#DisableDatabase(%s) DISALLOW %s ! %s", dbname, node.Host, err))
		}

		sq = fmt.Sprintf(`SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = '%s' AND pid <> pg_backend_pid()`, dbname)
		log.Trace(fmt.Sprintf(`RDPG#DisableDatabase(%s) TERMINATE %s > %s`, dbname, node.Host, sq))
		_, err = db.Exec(sq)
		if err != nil {
			log.Error(fmt.Sprintf("RDPG#DisableDatabase(%s) TERMINATE %s ! %s", dbname, node.Host, err))
		}

		sq = fmt.Sprintf(`ALTER DATABASE %s OWNER TO %s`, dbname, node.User)
		log.Trace(fmt.Sprintf(`RDPG#DisableDatabase(%s) OWNER %s > %s`, dbname, node.Host, sq))
		_, err = db.Exec(sq)
		if err != nil {
			log.Error(fmt.Sprintf("RDPG#DisableDatabase(%s) OWNER %s ! %s", dbname, node.Host, err))
		}
		db.Close()
	}

	return nil
}

func (r *RDPG) BackupDatabase(dbname string) (err error) {
	log.Error(fmt.Sprintf("RDPG#BackupDatabase(%s) TODO: IMPLEMENT", dbname))
	return nil
}

func (r *RDPG) DropDatabase(dbname string) (err error) {
	for _, node := range r.Nodes() {
		node.Database = "postgres"
		db, err := node.Connect()
		if err != nil {
			log.Error(fmt.Sprintf("RDPG#DropDatabase(%s) %s ! %s", dbname, node.Host, err))
			return err
		}

		// sq := fmt.Sprintf(SELECT slot_name FROM pg_replication_slots WHERE database='%s',dbname);
		// pg_recvlogical --drop-slot

		// TODO: How do we drop a database in bdr properly?
		sq := fmt.Sprintf(`DROP DATABASE IF EXISTS %s`, dbname)
		log.Trace(fmt.Sprintf(`RDPG#DropDatabase(%s) %s > %s`, dbname, node.Host, sq))
		_, err = db.Exec(sq)
		if err != nil {
			log.Error(fmt.Sprintf("RDPG#DropDatabase(%s) %s ! %s", dbname, node.Host, err))
		}
		db.Close()
	}
	return nil
}

func (r *RDPG) DropUser(name string) (err error) {
	for _, node := range r.Nodes() {
		node.Database = "postgres"
		db, err := node.Connect()
		if err != nil {
			log.Error(fmt.Sprintf("RDPG#DropUser(%s) %s ! %s", name, node.Host, err))
			return err
		}

		sq := fmt.Sprintf(`DROP USER %s`, name)
		log.Trace(fmt.Sprintf(`RDPG#DropUser(%s) %s > %s`, name, node.Host, sq))
		_, err = db.Exec(sq)
		if err != nil {
			log.Error(fmt.Sprintf("RDPG#DropUser(%s) %s ! %s", name, node.Host, err))
		}
		db.Close()
	}
	return nil
}
