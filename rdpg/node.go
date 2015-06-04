package rdpg

import (
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/wayneeseguin/rdpg-agent/log"
)

type Node struct {
	Host        string `db:"host" json:"host"`
	Port        string `db:"port" json:"port"`
	User        string `db:"user" json:"user"`
	Database    string `db:"database" json:"database"`
	Name        string `db:"node_name" json:"node_name" json:"name"`
	LocalDSN    string `db:"node_local_dsn" json:"node_local_dsn" json:"local_dsn"`
	InitFromDSN string `db:"node_init_from_dsn" json:"node_init_from_dsn" json:"init_from_dsn"`
}

func NewNode(host, port, user, database string) Node {
	return Node{Host: host, Port: port, User: user, Database: database}
}

func (n *Node) Connect() (db *sqlx.DB, err error) {
	uri := n.URI()
	db, err = sqlx.Connect("postgres", uri)
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.Node#Connect(): %s:\n%s\n", uri, err))
		return db, err
	}
	return db, nil
}

func UserExists() {
	//"SELECT rolname FROM pg_roles WHERE rolname='${rdpgUser}';"
}

func DatabaseExists() {
	//"SELECT datname FROM pg_database WHERE datname='${rdpgDB}';"
}

func CreateDatabase() {
	//
}

func CreateExtension(name string) {
	//("CREATE EXTENSION IF NOT EXISTS ?;",name)
}

func (n *Node) URI() (uri string) {
	d := "postgres://%s@%s:%s/%s?fallback_application_name=%s&connect_timeout=%s&sslmode=%s"
	uri = fmt.Sprintf(d, n.User, n.Host, n.Port, n.Database, "rdpg-agent", "5", "disable")
	return
}

func (n *Node) CreateDatabase(name string) (err error) {
	db, err := n.Connect()
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.Node#CreateDatabase(): %s\n", err))
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
		return errors.New(fmt.Sprintf("Node user is not postgres, can not create a user with '%s'", n.User))
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
	_, err = db.Exec("ALTER USER ? ENCRYPTED PASSWORD '?';", name, password)
	if err != nil {
		return err
	}
	log.Debug(fmt.Sprintf("Set password for user: ?\n", name))

	return nil
}
