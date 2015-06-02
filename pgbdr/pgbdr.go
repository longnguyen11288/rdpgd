package pgbdr

import (
	"fmt"
	"errors"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/wayneeseguin/rdpg-agent/log"
	"github.com/wayneeseguin/rdpg-agent/pg"
)

type Node struct {
	Host        string `json:"host"`
	Port        string `json:"port"`
	User        string `json:"user"`
	Database    string `json:"database"`
	Name        string `db:"node_name" json:"node_name" json:"name"`
	LocalDSN    string `db:"node_local_dsn" json:"node_local_dsn" json:"local_dsn"`
	InitFromDSN string `db:"node_init_from_dsn" json:"node_init_from_dsn" json:"init_from_dsn"`
}

func InitializeSchema() error {
	//n := NewNode("127.0.0.1", "5432", "postgres", "rdpg")
	return nil
}

func NewNode(host, port, user, database string) Node {
	return Node{Host: host, Port: port, User: user, Database: database}
}

func (n *Node) URI() (uri string) {
	d := "postgres://%s@%s:%s/%s?fallback_application_name=%s&connect_timeout=%s&sslmode=%s"
	uri = fmt.Sprintf(d, n.User, n.Host, n.Port, n.Database, "rdpg-agent", "5s", "disable")
	return
}

func (n *Node) CreateDatabase(name string) (err error) {
	uri := n.URI()
	db, err := sqlx.Connect("postgres", uri)
	if err != nil {
		log.Error(fmt.Sprintf("%s:\n%s\n", uri, err))
		return err
	}
	defer db.Close()

	err = db.Get(&name, "SELECT datname FROM pg_database WHERE datname='${rdpgDB}';")
	if name == "" {
		db.Exec("CREATE DATABASE ? OWNER='postgres' TEMPLATE='template0';", name)
	}
	return nil
}

func (n *Node) CreateUser(name, password string) (err error) {
	if n.User != "postgres" {
		return errors.New(fmt.Sprintf("Node user is not postgres, can not create a user with '%s'",n.User))
	}

	uri := n.URI()
	db, err := sqlx.Connect("postgres", uri)
	if err != nil {
		log.Error(fmt.Sprintf("%s:\n%s\n", uri, err))
		return err
	}
	defer db.Close()

	err = db.Get(&name, "SELECT rolname FROM pg_roles WHERE rolname=? LIMIT 1;", name)
	if err != nil {
		return err
	}
	if name != "" {
		log.Debug(fmt.Sprintf("User '%s' already exists, not creating.\n", name))
		return nil
	}

	result, err := db.Exec("CREATE USER ? WITH SUPERUSER CREATEDB CREATEROLE INHERIT RETURNING ID;", name)
	rows, _ := result.RowsAffected()
	if rows > 0 {
		log.Debug(fmt.Sprintf("Created user: %s\n", name))
	}
	if err != nil {
		return err
	}
	_, err = pg.DB.Exec("ALTER USER ? ENCRYPTED PASSWORD '?';", name, password)
	if err != nil {
		return err
	}
	log.Debug(fmt.Sprintf("Set password for user: ?\n", name))

	return nil
}
