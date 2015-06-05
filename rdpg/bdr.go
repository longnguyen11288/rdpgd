package rdpg

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/wayneeseguin/rdpg-agent/log"
)

// TODO: RDPG Struct => RDPG Struct, allowing for multiple instances of RDPG
func NewRDPG(uri string) *RDPG {
	if uri == "" || uri[0:13] != "postgresql://" {
		log.Error(fmt.Sprintf("NewRDPG() uri malformed: '%s'", uri))
		return nil
	}
	return &RDPG{URI: uri}
}

func (r *RDPG) Nodes() (nodes []Node) {
	db, err := r.connect()
	if err != nil {
		log.Error(fmt.Sprintf("Nodes() %s", err))
	}
	rows, err := db.Query("SELECT host,port,user,'postgres' FROM rdpg.nodes;")
	if err != nil {
		log.Error(fmt.Sprintf("Nodes() %s", err))
	} else {
		sqlx.StructScan(rows, nodes)
	}
	return nodes
}

func (r *RDPG) CreateUser(username, password string) error {
	for _, node := range r.Nodes() {
		db, err := node.Connect()
		if err != nil {
			log.Error(fmt.Sprintf("CreateUser() %s", err))
			return err
		}
		if _, err := db.Exec("CREATE USER ?;", username); err != nil {
			log.Error(fmt.Sprintf("CreateUser() %s", err))
		}
		if _, err := db.Exec("ALTER USER ? ENCRYPTED PASSWORD '?';", username, password); err != nil {
			log.Error(fmt.Sprintf("CreateUser() %s", err))
		}
		db.Close()
	}
	return nil
}

func (r *RDPG) CreateDatabase(name, owner string) error {
	// TODO: Drop Database on all nodes if err != nil for any operation below
	for _, node := range r.Nodes() {
		db, err := node.Connect()
		if err != nil {
			log.Error(fmt.Sprintf("CreateDatabase() %s", err))
			return err
		}

		if _, err := db.Exec(`CREATE DATABASE ? OWNER ? TEMPLATE='template0'`, name, owner); err != nil {
			log.Error(fmt.Sprintf("CreateDatabase() %s", err))
			return err
		}

		if _, err := db.Exec(`CREATE EXTENSION IF NOT EXISTS btree_gist`); err != nil {
			log.Error(fmt.Sprintf("CreateDatabase() %s", err))
			return err
		}

		if _, err := db.Exec(`CREATE EXTENSION IF NOT EXISTS bdr`); err != nil {
			log.Error(fmt.Sprintf("CreateDatabase() %s", err))
			return err
		}
		db.Close()
	}
	return nil
}

func (r *RDPG) CreateReplicationGroup(dbname string) error {
	nodes := r.Nodes()
	// TODO: Drop Database on all nodes if err != nil for any operation below
	for index, node := range nodes {
		db, err := node.Connect()
		if err != nil {
			log.Error(fmt.Sprintf("CreateReplicationGroup() %s", err))
			return err
		}
		if index == 0 {
			sq := `SELECT bdr.bdr_group_create(
				local_node_name := '${vmName}${vmIndex}',
				node_external_dsn := 'host=${nodes[0]} port=${pgbdrPort} user=${rdpgUser} dbname=${rdpgDB}'
			); `
			db.Exec(sq, node.Name, node.Host, node.Port, node.User, dbname)
		} else {
			sq := `SELECT bdr.bdr_group_join(
				local_node_name := '${vmName}${vmIndex}',
				node_external_dsn := 'host=${nodes[${vmIndex}]} port=${pgbdrPort} user=${rdpgUser} dbname=${rdpgDB}',
				join_using_dsn := 'host=${nodes[0]} port=${pgbdrPort} user=${rdpgUser} dbname=${rdpgDB}'
			); `
			db.Exec(sq,
				nodes[0].Name, nodes[0].Host, nodes[0].Port, nodes[0].User, dbname,
				node.Host, node.Port, node.User, dbname)
		}
		db.Close()
	}
	// SELECT bdr.bdr_node_join_wait_for_ready();
	return nil
}

func (r *RDPG) connect() (dr *sqlx.DB, err error) {
	db, err := sqlx.Connect("postgres", r.URI)
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.Node#Connect(): %s:\n%s\n", r.URI, err))
		return db, err
	}
	return db, nil
}

func (r *RDPG) DisableDatabase() (err error) {
	return nil
}

func (r *RDPG) BackupDatabase() (err error) {
	return nil
}

func (r *RDPG) DeleteDatabase() (err error) {
	return nil
}

func (r *RDPG) DeleteUser() (err error) {
	return nil
}
