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

type Host struct {
	Host        string `db:"host" json:"host"`
	Port        string `db:"port" json:"port"`
	User        string `db:"user" json:"user"`
	Database    string `db:"database" json:"database"`
	Name        string `db:"node_name" json:"name"`
	LocalDSN    string `db:"node_local_dsn" json:"node_local_dsn"`
	InitFromDSN string `db:"node_init_from_dsn" json:"node_init_from_dsn"`
}

func NewHost(host, port, user, database string) Host {
	return Host{Host: host, Port: port, User: user, Database: database}
}

func (n *Host) Connect() (db *sqlx.DB, err error) {
	uri := n.URI()
	db, err = sqlx.Connect(`postgres`, uri) // n.LocalDSN)
	if err != nil {
		log.Error(fmt.Sprintf(`rdpg.Host#Connect() %s ! %s`, uri, err))
		return db, err
	}
	return db, nil
}

func (n *Host) AdminAPI(method, path string) (err error) {
	url := fmt.Sprintf("http://%s:%s/%s", n.Host, os.Getenv("RDPG_ADMIN_PORT"), path)
	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(`{}`)))
	// req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(os.Getenv("RDPG_ADMIN_USER"), os.Getenv("RDPG_ADMIN_PASS"))
	client := &http.Client{}
	log.Trace(fmt.Sprintf(`Host#AdminAPI(%s,%s) %s`, method, path, url))
	resp, err := client.Do(req)
	if err != nil {
		log.Error(fmt.Sprintf(`Host#AdminAPI(%s,%s) ! %s`, method, url, err))
	}
	resp.Body.Close()

	return
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

func (n *Host) URI() (uri string) {
	d := `postgres://%s@%s:%s/%s?fallback_application_name=%s&connect_timeout=%s&sslmode=%s`
	uri = fmt.Sprintf(d, n.User, n.Host, n.Port, n.Database, `rdpg-agent`, `5`, `disable`)
	return
}

// Create a given user on a single target host.
func (n *Host) CreateUser(name, password string) (err error) {
	if n.User != `postgres` {
		return errors.New(fmt.Sprintf(`Host user is not postgres, can not create a user with '%s'`, n.User))
	}

	uri := n.URI()
	db, err := sqlx.Connect(`postgres`, uri)
	if err != nil {
		log.Error(fmt.Sprintf(`Host#CreateUser(%s) %s ! %s`, name, uri, err))
		return err
	}
	defer db.Close()

	err = db.Get(&name, `SELECT rolname FROM pg_roles WHERE rolname=? LIMIT 1;`, name)
	if err != nil {
		log.Error(fmt.Sprintf(`Host#CreateUser(%s) %s ! %s`, name, n.Host, err))
		return err
	}
	if name != "" {
		log.Debug(fmt.Sprintf(`User %s already exists, skipping.`, name))
		return nil
	}

	sq := fmt.Sprintf(`CREATE USER %s WITH SUPERUSER CREATEDB CREATEROLE INHERIT;`, name)
	result, err := db.Exec(sq)
	rows, _ := result.RowsAffected()
	if rows > 0 {
		log.Debug(fmt.Sprintf(`Host#CreateUser(%s) %s User Created`, n.Host, name))
	}
	if err != nil {
		log.Error(fmt.Sprintf(`Host#CreateUser(%s) %s ! %s`, name, n.Host, err))
		return err
	}
	sq = fmt.Sprintf(`ALTER USER %s ENCRYPTED PASSWORD %s;`, name, password)
	_, err = db.Exec(sq)
	if err != nil {
		log.Error(fmt.Sprintf(`Host#CreateUser(%s) %s ! %s`, name, n.Host, err))
		return err
	}

	return nil
}

// Create a given database and owner on a single target host.
func (n *Host) CreateDatabase(dbname, owner string) (err error) {
	n.Database = "postgres"
	db, err := n.Connect()
	if err != nil {
		log.Error(fmt.Sprintf("Host#CreateDatabase(%s) %s ! %s", dbname, n.Host, err))
		return
	}
	defer db.Close()

	sq := fmt.Sprintf(`CREATE DATABASE %s WITH OWNER %s TEMPLATE template0 ENCODING 'UTF8'`, dbname, owner)
	log.Trace(fmt.Sprintf(`Host#CreateDatabase(%s) %s > %s`, dbname, n.Host, sq))
	_, err = db.Query(sq)
	if err != nil {
		log.Error(fmt.Sprintf("Host#CreateDatabase(%s) %s ! %s", dbname, n.Host, err))
		return
	}

	sq = fmt.Sprintf(`REVOKE ALL ON DATABASE "%s" FROM public`, dbname)
	log.Trace(fmt.Sprintf(`Host#CreateDatabase(%s) %s > %s`, dbname, n.Host, sq))
	_, err = db.Exec(sq)
	if err != nil {
		log.Error(fmt.Sprintf("Host#CreateDatabase(%s) %s ! %s", dbname, n.Host, err))
	}

	sq = fmt.Sprintf(`GRANT ALL PRIVILEGES ON DATABASE %s TO %s`, dbname, owner)
	log.Trace(fmt.Sprintf(`Host#CreateDatabase(%s) %s > %s`, dbname, n.Host, sq))
	_, err = db.Query(sq)
	if err != nil {
		log.Error(fmt.Sprintf(`Host#CreateDatabase(%s) %s ! %s`, dbname, n.Host, err))
		return
	}
	return nil
}

// Create extensions on a single target host.
func (n *Host) CreateExtensions(exts []string) (err error) {
	db, err := n.Connect()
	if err != nil {
		log.Error(fmt.Sprintf("Host#CreateExtensions() %s ! %s", n.URI(), err))
		return
	}
	defer db.Close()

	_, err = db.Query(`CREATE EXTENSION IF NOT EXISTS btree_gist`)
	if err != nil {
		log.Error(fmt.Sprintf("Host#CreateExtensions() %s ! %s", n.Host, err))
		return
	}

	_, err = db.Query(`CREATE EXTENSION IF NOT EXISTS bdr`)
	if err != nil {
		log.Error(fmt.Sprintf("Host#CreateExtensions() %s ! %s", n.Host, err))
		return
	}
	return
}
