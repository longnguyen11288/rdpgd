package rdpg

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/wayneeseguin/rdpg-agent/log"
)

type BDR struct {
	URI string
}

func NewBDR(uri string) *BDR {
	if uri == "" || uri[0:13] != "postgresql://" {
		log.Error(fmt.Sprintf("NewBDR() uri malformed: '%s'", uri))
		return nil
	}
	return &BDR{URI: uri}
}

func (b *BDR) Nodes() (nodes []Node) {
	db, err := b.connect()
	if err != nil {
		log.Error(fmt.Sprintf("Nodes() %s", err))
	}
	rows,err := db.Query("SELECT host,port,user,'postgres' FROM rdpg.nodes;")
	if err != nil {
		log.Error(fmt.Sprintf("Nodes() %s", err))
	} else {
		sqlx.StructScan(rows, nodes)
	}
	return nodes
}

func (b *BDR) CreateUser(username,password string) error {
	for _, node := range b.Nodes() {
		db, err := node.Connect()
		if err != nil {
			log.Error(fmt.Sprintf("CreateUser() %s", err))
			return err
		}
		db.Exec("CREATE USER ? WITH SUPERUSER CREATEDB CREATEROLE INHERIT;",username)
		db.Exec("ALTER USER ? ENCRYPTED PASSWORD '?';",username,password)
	}
	return nil
}

func (b *BDR) CreateDatabase(name, owner string) error {
	for _, node := range b.Nodes() {
		db, err := node.Connect()
		if err != nil {
			log.Error(fmt.Sprintf("CreateDatabase() %s", err))
			return err
		}
		db.Exec(`CREATE DATABASE ? OWNER ? TEMPLATE='template0'`, name, owner)
		db.Exec(`CREATE EXTENSION IF NOT EXISTS btree_gist`)
		db.Exec(`CREATE EXTENSION IF NOT EXISTS bdr`)
	}
	return nil
}

func (b *BDR) CreateReplicationGroup(dbname string) error {
	nodes := b.Nodes()
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
	}
	// SELECT bdr.bdr_node_join_wait_for_ready();
	return nil
}

func (b *BDR) connect() (db *sqlx.DB, err error) {
	db, err = sqlx.Connect("postgres", b.URI)
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.Node#Connect(): %s:\n%s\n", b.URI, err))
		return db, err
	}
	return db, nil
}
