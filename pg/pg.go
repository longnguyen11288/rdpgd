package pg

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/wayneeseguin/rdpgd/log"
)

type PG struct {
	// Name string `` ???
	IP             string `db:"ip" json:"ip"`
	Port           string `db:"port" json:"port"`
	User           string `db:"user" json:"user"`
	Pass           string `db:"pass" json:"pass"`
	Database       string `db:"database" json:"database"`
	ConnectTimeout string `db:"connect_timeout" json:"connect_timeout,omitempty"`
	SSLMode        string `db:"sslmode" json:"sslmode,omitempty"`
	URI            string `db:"uri" json:"uri"`
	DSN            string `db:"ds" json:"dsn"`
}

// Create and return a new PG using default parameters
func NewPG(host, port, user, database, pass string) (p *PG) {

	p = &PG{IP: host, Port: port, User: user, Database: database, Pass: pass}

	p.ConnectTimeout = `5` // Default connection time out.
	p.SSLMode = `disable`  // Default disable SSL Mode, can be overwritten using Set()

	p.pgURI()
	p.pgDSN()

	log.Trace(fmt.Sprintf(`pg.PG#NewPG() New PG struct: %+v`, p))
	return
}

// Check if the given PostgreSQL User Exists on the host.
func (p *PG) UserExists(dbuser string) (exists bool, err error) {
	p.Set(`database`, `postgres`)
	db, err := p.Connect()
	if err != nil {
		log.Error(fmt.Sprintf("pg.PG<%s>#UserExists(%s) %s ! %s", p.IP, dbuser, p.URI, err))
		return
	}
	defer db.Close()

	type name struct {
		Name string `db:"name"`
	}
	var n name
	sq := fmt.Sprintf(`SELECT rolname AS name FROM pg_roles WHERE rolname='%s' LIMIT 1;`, dbuser)
	err = db.Get(&n, sq)
	if err != nil {
		if err == sql.ErrNoRows {
			exists = false
			err = nil
		} else {
			log.Error(fmt.Sprintf(`pg.PG<%s>#UserExists(%s) ! %s`, p.IP, dbuser, err))
			return
		}
	}
	if n.Name != "" {
		exists = true
	} else {
		exists = false
	}
	return
}

// Check if the given PostgreSQL Database Exists on the host.
func (p *PG) DatabaseExists(dbname string) (exists bool, err error) {
	p.Set(`database`, `postgres`)
	db, err := p.Connect()
	if err != nil {
		log.Error(fmt.Sprintf("pg.PG<%s>#DatabaseExists(%s) %s ! %s", p.IP, dbname, p.URI, err))
		return
	}
	defer db.Close()

	type name struct {
		Name string `db:"name"`
	}
	var n name
	sq := fmt.Sprintf(`SELECT datname AS name FROM pg_database WHERE datname='%s' LIMIT 1;`, dbname)
	err = db.Get(&n, sq)
	if err != nil {
		if err == sql.ErrNoRows {
			exists = false
			err = nil
		} else {
			log.Error(fmt.Sprintf(`pg.PG<%s>#DatabaseExists(%s) ! %s`, p.IP, dbname, err))
			return
		}
	}
	if n.Name != "" {
		exists = true
	} else {
		exists = false
	}
	return
}

// Create a given user on a single target host.
func (p *PG) CreateUser(dbuser, dbpass string) (err error) {
	p.Set(`database`, `postgres`)
	db, err := p.Connect()
	if err != nil {
		log.Error(fmt.Sprintf("pg.PG<%s>#CreateUser(%s) %s ! %s", p.IP, dbuser, p.URI, err))
		return
	}
	defer db.Close()

	exists, err := p.UserExists(dbuser)
	if err != nil {
		log.Error(fmt.Sprintf("pg.PG<%s>#CreateUser(%s) ! %s", p.IP, dbuser, err))
		return
	}
	if exists {
		log.Debug(fmt.Sprintf(`User %s already exists, skipping.`, dbuser))
		return nil
	}

	sq := fmt.Sprintf(`CREATE USER %s;`, dbuser)
	log.Trace(fmt.Sprintf(`pg.PG<%s>#CreateUser(%s) > %s`, p.IP, dbuser, sq))
	_, err = db.Exec(sq)
	if err != nil {
		log.Error(fmt.Sprintf("pg.PG<%s>#CreateUser(%s) ! %s", p.IP, dbuser, err))
		db.Close()
		return err
	}

	sq = fmt.Sprintf(`ALTER USER %s ENCRYPTED PASSWORD '%s';`, dbuser, dbpass)
	log.Trace(fmt.Sprintf(`pg.PG<%s>#CreateUser(%s)`, p.IP, dbuser))
	_, err = db.Exec(sq)
	if err != nil {
		log.Error(fmt.Sprintf(`pg.PG<%s>#CreateUser(%s) ! %s`, p.IP, dbuser, err))
		return
	}

	return
}

// Create a given user on a single target host.
func (p *PG) UserGrantPrivileges(dbuser string, priviliges []string) (err error) {
	p.Set(`database`, `postgres`)
	db, err := p.Connect()
	if err != nil {
		log.Error(fmt.Sprintf("pg.PG<%s>#UserGrantPrivileges(%s) %s ! %s", p.IP, dbuser, p.URI, err))
		return
	}
	defer db.Close()

	for _, priv := range priviliges {
		sq := fmt.Sprintf(`ALTER USER %s GRANT %s;`, dbuser, priv)
		log.Trace(fmt.Sprintf(`pg.PG<%s>#UserGrantPrivileges(%s) > %s`, p.IP, dbuser, sq))
		result, err := db.Exec(sq)
		rows, _ := result.RowsAffected()
		if rows > 0 {
			log.Trace(fmt.Sprintf(`pg.PG<%s>#CreateUser(%s) Successfully Created.`, p.IP, dbuser))
		}
		if err != nil {
			log.Error(fmt.Sprintf(`pg.PG<%s>#CreateUser(%s) ! %s`, p.IP, dbuser, err))
			return err
		}
	}
	return nil
}

// Create a given database owned by user on a single target host.
func (p *PG) CreateDatabase(dbname, dbuser string) (err error) {
	p.Set(`database`, `postgres`)
	db, err := p.Connect()
	if err != nil {
		log.Error(fmt.Sprintf("pg.PG<%s>#CreateDatabase(%s,%s) %s ! %s", p.IP, dbname, dbuser, p.URI, err))
		return
	}
	defer db.Close()

	exists, err := p.UserExists(dbuser)
	if err != nil {
		log.Error(fmt.Sprintf("pg.PG<%s>#CreateDatabase(%s,%s) ! %s", p.IP, dbname, dbuser, err))
		return
	}
	if !exists {
		err = fmt.Errorf(`User does not exist, ensure that postgres user '%s' exists first.`, dbuser)
		log.Error(fmt.Sprintf("pg.PG<%s>#CreateDatabase(%s,%s) ! %s", p.IP, dbname, dbuser, err))
		return
	}

	sq := fmt.Sprintf(`CREATE DATABASE %s WITH OWNER %s TEMPLATE template0 ENCODING 'UTF8'`, dbname, dbuser)
	log.Trace(fmt.Sprintf(`pg.PG<%s>#CreateDatabase(%s,%s) > %s`, p.IP, dbname, dbuser, sq))
	_, err = db.Query(sq)
	if err != nil {
		log.Error(fmt.Sprintf("pg.PG<%s>#CreateDatabase(%s,%s) ! %s", p.IP, dbname, dbuser, err))
		return
	}

	sq = fmt.Sprintf(`REVOKE ALL ON DATABASE "%s" FROM public`, dbname)
	log.Trace(fmt.Sprintf(`pg.PG<%s>#CreateDatabase(%s,%s) > %s`, p.IP, dbname, dbuser, sq))
	_, err = db.Exec(sq)
	if err != nil {
		log.Error(fmt.Sprintf("pg.PG<%s>#CreateDatabase(%s,%s) ! %s", p.IP, dbname, dbuser, err))
	}

	sq = fmt.Sprintf(`GRANT ALL PRIVILEGES ON DATABASE %s TO %s`, dbname, dbuser)
	log.Trace(fmt.Sprintf(`pg.PG<%s>#CreateDatabase(%s,%s) > %s`, p.IP, dbname, dbuser, sq))
	_, err = db.Query(sq)
	if err != nil {
		log.Error(fmt.Sprintf("pg.PG<%s>#CreateDatabase(%s,%s) ! %s", p.IP, dbname, dbuser, err))
		return
	}
	return nil
}

// Create given extensions on a single target host.
func (p *PG) CreateExtensions(dbname string, exts []string) (err error) {
	p.Set(`database`, dbname)
	db, err := p.Connect()
	if err != nil {
		log.Error(fmt.Sprintf("pg.PG<%s>#CreateExtensions(%s) %s ! %s", p.IP, dbname, p.URI, err))
		return
	}

	for _, ext := range exts {
		sq := fmt.Sprintf(`CREATE EXTENSION IF NOT EXISTS %s;`, ext)
		log.Trace(fmt.Sprintf(`pg.PG<%s>#CreateExtensions() > %s`, p.IP, sq))
		_, err = db.Exec(sq)
		if err != nil {
			db.Close()
			log.Error(fmt.Sprintf("pg.PG<%s>#CreateExtensions() %s ! %s", p.IP, ext, err))
			return
		}
	}
	db.Close()
	return
}

func (p *PG) DisableDatabase(dbname string) (err error) {
	p.Set(`database`, `postgres`)
	db, err := p.Connect()
	if err != nil {
		log.Error(fmt.Sprintf("pg.PG<%s>#DisableDatabase(%s) %s ! %s", p.IP, dbname, p.URI, err))
		return
	}
	defer db.Close()

	sq := fmt.Sprintf(`SELECT rdpg.bdr_disable_database('%s');`, dbname)
	log.Trace(fmt.Sprintf(`pg.PG<%s>#DisableDatabase(%s) DISABLE > %s`, p.IP, dbname, sq))
	_, err = db.Exec(sq)
	if err != nil {
		log.Error(fmt.Sprintf("p.PG<%s>#DisableDatabase(%s) DISABLE ! %s", p.IP, dbname, err))
	}
	return
}

func (p *PG) BDRGroupCreate(group, dbname string) (err error) {
	p.Set(`database`, dbname)
	db, err := p.Connect()
	if err != nil {
		return
	}
	defer db.Close()
	sq := fmt.Sprintf(`SELECT bdr.bdr_group_create( local_node_name := '%s',
			node_external_dsn := 'host=%s port=%s user=%s dbname=%s'); `,
		group, p.IP, p.Port, p.User, dbname,
	)
	log.Trace(fmt.Sprintf(`p.PG#CreateReplicationGroup(%s) %s > %s`, dbname, p.IP, sq))
	_, err = db.Exec(sq)
	if err == nil {
		sq = `SELECT bdr.bdr_node_join_wait_for_ready();`
		log.Trace(fmt.Sprintf(`p.PG#CreateReplicationGroup(%s) %s > %s`, dbname, p.IP, sq))
		_, err = db.Exec(sq)
	}
	db.Close()

	return
}

func (p *PG) BDRGroupJoin(group, dbname string, target PG) (err error) {
	p.Set(`database`, dbname)
	db, err := p.Connect()
	if err != nil {
		return
	}
	defer db.Close()
	sq := fmt.Sprintf(`SELECT bdr.bdr_group_join( local_node_name := '%s',
				node_external_dsn := 'host=%s port=%s user=%s dbname=%s',
				join_using_dsn := 'host=%s port=%s user=%s dbname=%s'); `,
		group, p.IP, p.Port, p.User, p.Database,
		target.IP, target.Port, target.User, dbname,
	)
	log.Trace(fmt.Sprintf(`p.PG#CreateReplicationGroup(%s) %s > %s`, dbname, p.IP, sq))
	_, err = db.Exec(sq)
	if err == nil {
		sq = `SELECT bdr.bdr_node_join_wait_for_ready();`
		log.Trace(fmt.Sprintf(`p.PG#CreateReplicationGroup(%s) %s > %s`, dbname, p.IP, sq))
		_, err = db.Exec(sq)
	}
	db.Close()
	return
}

func (p *PG) StopReplication(dbname string) (err error) {
	// TODO Finish this function
	p.Set(`database`, `postgres`)
	db, err := p.Connect()
	if err != nil {
		log.Error(fmt.Sprintf("pg.PG<%s>#DropDatabase(%s) %s ! %s", p.IP, dbname, p.URI, err))
		return
	}
	// sq := fmt.Sprintf(SELECT slot_name FROM pg_replication_slots WHERE database='%s',dbname);
	// pg_recvlogical --drop-slot

	defer db.Close()
	return
}

func (p *PG) DropDatabase(dbname string) (err error) {
	p.Set(`database`, `postgres`)
	db, err := p.Connect()
	if err != nil {
		log.Error(fmt.Sprintf("pg.PG<%s>#DropDatabase(%s) %s ! %s", p.IP, dbname, p.URI, err))
		return
	}
	defer db.Close()

	exists, err := p.DatabaseExists(dbname)
	if err != nil {
		log.Error(fmt.Sprintf("pg.PG<%s>#DropDatabase(%s) ! %s", p.IP, dbname, err))
		return
	}
	if !exists {
		log.Error(fmt.Sprintf("pg.PG<%s>#DropDatabase(%s) Database %s already does not exist.", p.IP, dbname, err))
		return
	}

	// TODO: How do we drop a database in bdr properly?
	sq := fmt.Sprintf(`DROP DATABASE IF EXISTS %s`, dbname)
	log.Trace(fmt.Sprintf(`p.PG#DropDatabase(%s) %s DROP > %s`, dbname, p.IP, sq))
	_, err = db.Exec(sq)
	if err != nil {
		log.Error(fmt.Sprintf("p.PG#DropDatabase(%s) DROP %s ! %s", dbname, p.IP, err))
		return
	}
	return
}

func (p *PG) DropUser(dbuser string) (err error) {
	p.Set(`database`, `postgres`)
	db, err := p.Connect()
	if err != nil {
		log.Error(fmt.Sprintf("pg.PG<%s>#DropUser(%s) %s ! %s", p.IP, dbuser, p.URI, err))
		return
	}
	defer db.Close()

	exists, err := p.UserExists(dbuser)
	if err != nil {
		log.Error(fmt.Sprintf("pg.PG<%s>#DropUser(%s) ! %s", p.IP, dbuser, err))
		return
	}
	if !exists {
		log.Error(fmt.Sprintf("pg.PG<%s>#DropUser(%s) User %s already does not exist.", p.IP, dbuser, err))
		return
	}

	// TODO: How do we drop a database in bdr properly?
	sq := fmt.Sprintf(`DROP USER %s`, dbuser)
	log.Trace(fmt.Sprintf(`p.PG#DropDatabase(%s) %s DROP > %s`, dbuser, p.IP, sq))
	_, err = db.Exec(sq)
	if err != nil {
		log.Error(fmt.Sprintf("p.PG#DropDatabase(%s) DROP %s ! %s", dbuser, p.IP, err))
		return
	}

	return
}

// Set host property to given value then regenerate the URI and DSN properties.
func (p *PG) Set(key, value string) (err error) {
	switch key {
	case "ip":
		p.IP = value
	case "port":
		p.Port = value
	case "user":
		p.User = value
	case "database":
		p.Database = value
	case "connect_timeout":
		p.ConnectTimeout = value
	case "sslmode":
		p.SSLMode = value
	case "pass":
	case "default": // A Bug
		err = fmt.Errorf(`Attempt to set unknown key %s to value %s for host %+v.`, key, value, *p)
		return err
	}
	p.pgURI()
	p.pgDSN()

	return
}

// Build and set the host's URI property according to the pattern:
//   postgres://user:password@ip:port/database?sslmode=&connect_timeout=&...
func (p *PG) pgURI() {
	p.URI = "postgres://"
	if p.User != "" {
		p.URI += p.User
	}
	if p.Pass != "" {
		p.URI += fmt.Sprintf(`:%s`, p.Pass)
	}
	if p.IP != "" {
		p.URI += fmt.Sprintf(`@%s`, p.IP)
	}
	if p.Port != "" {
		p.URI += fmt.Sprintf(`:%s`, p.Port)
	}
	if p.Database != "" {
		p.URI += fmt.Sprintf(`/%s`, p.Database)
	}
	p.URI += fmt.Sprintf(`?sslmode=%s&fallback_application_name=rdpgd`, p.SSLMode)
	if p.ConnectTimeout != "" {
		p.URI += fmt.Sprintf(`&connect_timeout=%s`, p.ConnectTimeout)
	}
	return
}

// Build and set the host's DSN property
func (p *PG) pgDSN() {
	p.DSN = ""
	if p.IP != "" {
		p.DSN += fmt.Sprintf(` host=%s`, p.IP)
	}
	if p.Port != "" {
		p.DSN += fmt.Sprintf(` port=%s`, p.Port)
	}
	if p.User != "" {
		p.DSN += fmt.Sprintf(` user=%s`, p.User)
	}
	if p.Pass != "" {
		p.DSN += fmt.Sprintf(` password=%s`, p.Pass)
	}
	if p.Database != "" {
		p.DSN += fmt.Sprintf(` dbname=%s`, p.Database)
	}
	if p.ConnectTimeout != "" {
		p.DSN += fmt.Sprintf(` connect_timeout=%s`, p.ConnectTimeout)
	}
	p.DSN += fmt.Sprintf(`fallback_application_name=rdpgd sslmode=%s`, p.SSLMode)
	return
}

// Connect to the host's database and return database connection object if successful
func (p *PG) Connect() (db *sqlx.DB, err error) {
	db, err = sqlx.Connect(`postgres`, p.URI)
	if err != nil {
		log.Error(fmt.Sprintf(`pg.PG<%s>#Connect() %s ! %s`, p.IP, p.URI, err))
		return db, err
	}
	return db, nil
}
