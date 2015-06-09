package rdpg

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"

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
	db, err = sqlx.Connect(`postgres`, uri) // n.LocalDSN)
	if err != nil {
		log.Error(fmt.Sprintf(`rdpg.Node#Connect(): %s :: %s`, uri, err))
		return db, err
	}
	return db, nil
}

func ExistsUser(user string) (exists bool, err error) {
	// As `postgres` user,
	// type name string
	// res,err := db.Select(&name, `SELECT rolname FROM pg_roles WHERE rolname='$1';`,user)
	// exits = (res.NumRows() > 0)
	return
}

func ExistsDatabase(dbname string) (exists bool, err error) {
	// As `postgres` user,
	// type name string
	//res,err := db.Select(&name,`SELECT datname FROM pg_database WHERE datname='$1';`,dbname)
	// exits = (res.NumRows() > 0)
	return
}

func CreateExtension(name string) (err error) {
	// sq := fmt.Sprintf(`CREATE EXTENSION IF NOT EXISTS %s;`,name)
	// _,err = db.Exec(sq)
	return
}

func (n *Node) URI() (uri string) {
	d := `postgres://%s@%s:%s/%s?fallback_application_name=%s&connect_timeout=%s&sslmode=%s`
	uri = fmt.Sprintf(d, n.User, n.Host, n.Port, n.Database, `rdpg-agent`, `5`, `disable`)
	return
}

func (n *Node) CreateUser(name, password string) (err error) {
	if n.User != `postgres` {
		return errors.New(fmt.Sprintf(`Node user is not postgres, can not create a user with '%s'`, n.User))
	}

	uri := n.URI()
	db, err := sqlx.Connect(`postgres`, uri)
	if err != nil {
		log.Error(fmt.Sprintf(`Node#CreateUser() %s %s`, uri, err))
		return err
	}
	defer db.Close()

	err = db.Get(&name, `SELECT rolname FROM pg_roles WHERE rolname=? LIMIT 1;`, name)
	if err != nil {
		log.Error(fmt.Sprintf(`Node#CreateUser() %s`, err))
		return err
	}
	if name != `` {
		log.Debug(fmt.Sprintf(`User '%s' already exists, not creating.`, name))
		return nil
	}

	sq := fmt.Sprintf(`CREATE USER %s WITH SUPERUSER CREATEDB CREATEROLE INHERIT;`, name)
	result, err := db.Exec(sq)
	rows, _ := result.RowsAffected()
	if rows > 0 {
		log.Debug(fmt.Sprintf(`Created user: %s`, name))
	}
	if err != nil {
		log.Error(fmt.Sprintf(`Node#CreateUser() %s`, err))
		return err
	}
	sq = fmt.Sprintf(`ALTER USER %s ENCRYPTED PASSWORD %s;`, name, password)
	_, err = db.Exec(sq)
	if err != nil {
		log.Error(fmt.Sprintf(`Node#CreateUser() %s`, err))
		return err
	}
	log.Debug(fmt.Sprintf(`Set password for user: ?`, name))

	return nil
}

func (n *Node) CreateDatabase(dbname, owner string) (err error) {
	n.Database = "postgres"
	db, err := n.Connect()
	if err != nil {
		log.Error(fmt.Sprintf("Node#CreateDatabase() error connecting to database %s", err))
		return
	}
	defer db.Close()

	sq := fmt.Sprintf(`CREATE DATABASE %s WITH OWNER %s TEMPLATE template0 ENCODING 'UTF8'`, dbname, owner)
	log.Trace(sq)
	_, err = db.Query(sq)
	if err != nil {
		log.Error(fmt.Sprintf("Node#CreateDatabase() %s", err))
		return
	}

	sq = fmt.Sprintf(`REVOKE ALL ON DATABASE "%s" FROM public`, dbname)
	log.Trace(sq)
	_, err = db.Exec(sq)
	if err != nil {
		log.Error(fmt.Sprintf("Node#CreateDatabase() %s", err))
	}

	sq = fmt.Sprintf(`GRANT ALL PRIVILEGES ON DATABASE %s TO %s`, dbname, owner)
	log.Trace(sq)
	_, err = db.Query(sq)
	if err != nil {
		log.Error(fmt.Sprintf("Node#CreateDatabase() %s", err))
		return
	}
	return nil
}

func (n *Node) CreateExtensions(exts []string) (err error) {
	db, err := n.Connect()
	if err != nil {
		log.Error(fmt.Sprintf("Node#CreateExtensions() %s :: %s", n.URI(), err))
		return
	}
	defer db.Close()

	_, err = db.Query(`CREATE EXTENSION IF NOT EXISTS btree_gist`)
	if err != nil {
		log.Error(fmt.Sprintf("Node#CreateExtensions() %s", err))
		return
	}

	_, err = db.Query(`CREATE EXTENSION IF NOT EXISTS bdr`)
	if err != nil {
		log.Error(fmt.Sprintf("Node#CreateExtensions() %s", err))
		return
	}
	return
}

func (n *Node) AdminAPI(method, path string) (err error) {
	url := fmt.Sprintf("http://%s:%s/%s", n.Host, os.Getenv("RDPG_ADMIN_PORT"), path)
	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(`{}`)))
	// req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(os.Getenv("RDPG_ADMIN_USER"), os.Getenv("RDPG_ADMIN_PASS"))
	client := &http.Client{}
	log.Trace(fmt.Sprintf(`Node#AdminAPI() %s %s`, method, url))
	resp, err := client.Do(req)
	if err != nil {
		log.Error(fmt.Sprintf(`Node#AdminAPI() %s %s :: %s`, method, url, err))
	}
	resp.Body.Close()

	return
}
