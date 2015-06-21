package pg

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/wayneeseguin/rdpgd/log"
)

type Host struct {
	// Name string `` ???
	IP             string `db:"ip" json:"ip"`
	Port           string `db:"port" json:"port"`
	User           string `db:"user" json:"user"`
	Database       string `db:"database" json:"database"`
	ConnectTimeout string `db:"connect_timeout" json:"connect_timeout,omitempty"`
	SSLMode        string `db:"sslmode" json:"sslmode,omitempty"`
	URI            string `db:"uri" json:"uri"`
	DSN            string `db:"ds" json:"dsn"`
}

// Create and return a new host using default parameters
func NewHost(host, port, user, database string) (h Host) {
	h = Host{IP: host, Port: port, User: user, Database: database}
	return
}

// Check if the given PostgreSQL User Exists on the host.
func (h *Host) PGUserExists(dbuser string) (exists bool, err error) {
	h.Set(`database`, `postgres`)
	db, err := h.PGConnect()
	if err != nil {
		log.Error(fmt.Sprintf("pg.Host<%s>#PGUserExists(%s) %s ! %s", h.IP, dbuser, h.URI, err))
		return
	}
	defer db.Close()

	var name string
	err = db.Get(&name, `SELECT rolname FROM pg_roles WHERE rolname=? LIMIT 1;`, dbuser)
	if err != nil {
		log.Error(fmt.Sprintf(`pg.Host<%s>#PGUserExists(%s) ! %s`, h.IP, dbuser, err))
		return
	}
	if name != "" {
		exists = true
	} else {
		exists = false
	}
	return
}

// Check if the given PostgreSQL Database Exists on the host.
func (h *Host) PGDatabaseExists(dbname string) (exists bool, err error) {
	h.Set(`database`, `postgres`)
	db, err := h.PGConnect()
	if err != nil {
		log.Error(fmt.Sprintf("pg.Host<%s>#PGDatabaseExists(%s) %s ! %s", h.IP, dbname, h.URI, err))
		return
	}
	defer db.Close()

	var name string
	err = db.Get(&name, `SELECT datname FROM pg_database WHERE datname=?;`, dbname)
	if err != nil {
		log.Error(fmt.Sprintf(`pg.Host<%s>#PGDatabaseExists(%s) ! %s`, h.IP, dbname, err))
		return
	}
	if name != "" {
		exists = true
	} else {
		exists = false
	}
	return
}

// Create a given user on a single target host.
func (h *Host) PGCreateUser(dbuser, dbpass string) (err error) {
	h.Set(`database`, `postgres`)
	db, err := h.PGConnect()
	if err != nil {
		log.Error(fmt.Sprintf("pg.Host<%s>#PGCreateUser(%s) %s ! %s", h.IP, dbuser, h.URI, err))
		return
	}
	defer db.Close()

	exists, err := h.PGUserExists(dbuser)
	if err != nil {
		log.Error(fmt.Sprintf("pg.Host<%s>#CreateUser(%s) ! %s", h.IP, dbuser, err))
		return
	}
	if exists {
		log.Debug(fmt.Sprintf(`User %s already exists, skipping.`, dbuser))
		return nil
	}

	sq := fmt.Sprintf(`CREATE USER %s;`, dbuser)
	log.Trace(fmt.Sprintf(`pg.Host<%s>#CreateUser(%s) > %s`, h.IP, dbuser, sq))
	_, err = db.Exec(sq)
	if err != nil {
		log.Error(fmt.Sprintf("pg.Host<%s>#CreateUser(%s) ! %s", h.IP, dbuser, err))
		db.Close()
		return err
	}

	sq = fmt.Sprintf(`ALTER USER %s ENCRYPTED PASSWORD %s;`, dbuser, dbpass)
	log.Trace(fmt.Sprintf(`pg.Host<%s>#PGCreateUser(%s)`, h.IP, dbuser))
	_, err = db.Exec(sq)
	if err != nil {
		log.Error(fmt.Sprintf(`pg.Host<%s>#PGCreateUser(%s) ! %s`, h.IP, dbuser, err))
		return
	}

	return
}

// Create a given user on a single target host.
func (h *Host) PGUserGrantPrivileges(dbuser string, priviliges []string) (err error) {
	h.Set(`database`, `postgres`)
	db, err := h.PGConnect()
	if err != nil {
		log.Error(fmt.Sprintf("pg.Host<%s>#PGUserGrantPrivileges(%s) %s ! %s", h.IP, dbuser, h.URI, err))
		return
	}
	defer db.Close()

	for _, priv := range priviliges {
		sq := fmt.Sprintf(`ALTER USER %s GRANT %s;`, dbuser, priv)
		log.Trace(fmt.Sprintf(`pg.Host<%s>#PGUserGrantPrivileges(%s) %s > %s`, h.IP, dbuser, sq))
		result, err := db.Exec(sq)
		rows, _ := result.RowsAffected()
		if rows > 0 {
			log.Trace(fmt.Sprintf(`pg.Host<%s>#CreateUser(%s) Successfully Created.`, h.IP, dbuser))
		}
		if err != nil {
			log.Error(fmt.Sprintf(`pg.Host<%s>#CreateUser(%s) ! %s`, h.IP, dbuser, err))
			return err
		}
	}
	return nil
}

// Create a given database owned by user on a single target host.
func (h *Host) PGCreateDatabase(dbname, dbuser string) (err error) {
	h.Set(`database`, `postgres`)
	db, err := h.PGConnect()
	if err != nil {
		log.Error(fmt.Sprintf("pg.Host<%s>#CreateDatabase(%s,%s) %s ! %s", h.IP, dbname, dbuser, h.URI, err))
		return
	}
	defer db.Close()

	exists, err := h.PGUserExists(dbuser)
	if err != nil {
		log.Error(fmt.Sprintf("pg.Host<%s>#CreateDatabase(%s,%s) ! %s", h.IP, dbname, dbuser, err))
		return
	}
	if !exists {
		err = fmt.Errorf(`User does not exist, ensure that postgres user '%s' exists first.`, dbuser)
		log.Error(fmt.Sprintf("pg.Host<%s>#CreateDatabase(%s,%s) ! %s", h.IP, dbname, dbuser, err))
		return
	}

	sq := fmt.Sprintf(`CREATE DATABASE %s WITH OWNER %s TEMPLATE template0 ENCODING 'UTF8'`, dbname, dbuser)
	log.Trace(fmt.Sprintf(`pg.Host<%s>#CreateDatabase(%s,%s) > %s`, h.IP, dbname, dbuser, sq))
	_, err = db.Query(sq)
	if err != nil {
		log.Error(fmt.Sprintf("pg.Host<%s>#CreateDatabase(%s,%s) ! %s", h.IP, dbname, dbuser, err))
		return
	}

	sq = fmt.Sprintf(`REVOKE ALL ON DATABASE "%s" FROM public`, dbname)
	log.Trace(fmt.Sprintf(`pg.Host<%s>#CreateDatabase(%s,%s) > %s`, h.IP, dbname, dbuser, sq))
	_, err = db.Exec(sq)
	if err != nil {
		log.Error(fmt.Sprintf("pg.Host<%s>#CreateDatabase(%s,%s) ! %s", h.IP, dbname, dbuser, err))
	}

	sq = fmt.Sprintf(`GRANT ALL PRIVILEGES ON DATABASE %s TO %s`, dbname, dbuser)
	log.Trace(fmt.Sprintf(`pg.Host<%s>#CreateDatabase(%s,%s) > %s`, h.IP, dbname, dbuser, sq))
	_, err = db.Query(sq)
	if err != nil {
		log.Error(fmt.Sprintf("pg.Host<%s>#CreateDatabase(%s,%s) ! %s", h.IP, dbname, dbuser, err))
		return
	}
	return nil
}

// Create given extensions on a single target host.
func (h *Host) PGCreateExtensions(dbname string, exts []string) (err error) {
	h.Set(`database`, dbname)
	db, err := h.PGConnect()
	if err != nil {
		log.Error(fmt.Sprintf("pg.Host<%s>#CreateExtensions(%s) %s ! %s", h.IP, dbname, h.URI, err))
		return
	}

	for _, ext := range exts {
		sq := fmt.Sprintf(`CREATE EXTENSION IF NOT EXISTS %s;`, ext)
		log.Trace(fmt.Sprintf(`pg.Host<%s>#CreateExtensions() > %s`, h.IP, sq))
		_, err = db.Exec(sq)
		if err != nil {
			db.Close()
			log.Error(fmt.Sprintf("pg.Host<%s>#PGCreateExtension() %s ! %s", h.IP, ext, err))
			return
		}
	}
	db.Close()
	return
}

func (h *Host) PGDisableDatabase(dbname string) (err error) {
	h.Set(`database`, `postgres`)
	db, err := h.PGConnect()
	if err != nil {
		log.Error(fmt.Sprintf("pg.Host<%s>#PGDisableDatabase(%s,%s) %s ! %s", h.IP, dbname, h.URI, err))
		return
	}
	defer db.Close()

	sq := fmt.Sprintf(`SELECT rdpg.bdr_disable_database('%s');`, dbname)
	log.Trace(fmt.Sprintf(`pg.Host<%s>#PGDisableDatabase(%s) DISABLE %s > %s`, dbname, h.IP, sq))
	_, err = db.Exec(sq)
	if err != nil {
		log.Error(fmt.Sprintf("RDPG#PGDisableDatabase(%s) DISABLE %s ! %s", dbname, h.IP, err))
	}

	return
}

func (h *Host) PGBDRGroupCreate(group, dbname string) (err error) {
	h.Set(`database`, dbname)
	db, err := h.PGConnect()
	if err != nil {
		return
	}
	defer db.Close()
	sq := fmt.Sprintf(`SELECT bdr.bdr_group_create( local_node_name := '%s',
			node_external_dsn := 'host=%s port=%s user=%s dbname=%s'); `,
		group, h.IP, h.Port, h.User, dbname,
	)
	log.Trace(fmt.Sprintf(`RDPG#CreateReplicationGroup(%s) %s > %s`, dbname, h.IP, sq))
	_, err = db.Exec(sq)
	if err == nil {
		sq = `SELECT bdr.bdr_node_join_wait_for_ready();`
		log.Trace(fmt.Sprintf(`RDPG#CreateReplicationGroup(%s) %s > %s`, dbname, h.IP, sq))
		_, err = db.Exec(sq)
	}
	db.Close()

	return
}

func (h *Host) PGBDRGroupJoin(group, dbname string, target Host) (err error) {
	h.Set(`database`, dbname)
	db, err := h.PGConnect()
	if err != nil {
		return
	}
	defer db.Close()
	sq := fmt.Sprintf(`SELECT bdr.bdr_group_join( local_node_name := '%s',
				node_external_dsn := 'host=%s port=%s user=%s dbname=%s',
				join_using_dsn := 'host=%s port=%s user=%s dbname=%s'); `,
		group, h.IP, h.Port, h.User, h.Database,
		target.IP, target.Port, target.User, dbname,
	)
	log.Trace(fmt.Sprintf(`RDPG#CreateReplicationGroup(%s) %s > %s`, dbname, h.IP, sq))
	_, err = db.Exec(sq)
	if err == nil {
		sq = `SELECT bdr.bdr_node_join_wait_for_ready();`
		log.Trace(fmt.Sprintf(`RDPG#CreateReplicationGroup(%s) %s > %s`, dbname, h.IP, sq))
		_, err = db.Exec(sq)
	}
	db.Close()
	return
}

func (h *Host) PGStopReplication(dbname string) (err error) {
	// TODO Finish this function
	h.Set(`database`, `postgres`)
	db, err := h.PGConnect()
	if err != nil {
		log.Error(fmt.Sprintf("pg.Host<%s>#PGDropDatabase(%s) %s ! %s", h.IP, dbname, h.URI, err))
		return
	}
	// sq := fmt.Sprintf(SELECT slot_name FROM pg_replication_slots WHERE database='%s',dbname);
	// pg_recvlogical --drop-slot

	defer db.Close()
	return
}

func (h *Host) PGDropDatabase(dbname string) (err error) {
	h.Set(`database`, `postgres`)
	db, err := h.PGConnect()
	if err != nil {
		log.Error(fmt.Sprintf("pg.Host<%s>#PGDropDatabase(%s) %s ! %s", h.IP, dbname, h.URI, err))
		return
	}
	defer db.Close()

	exists, err := h.PGDatabaseExists(dbname)
	if err != nil {
		log.Error(fmt.Sprintf("pg.Host<%s>#PGDropDatabase(%s) ! %s", h.IP, dbname, err))
		return
	}
	if !exists {
		log.Error(fmt.Sprintf("pg.Host<%s>#PGDropDatabase(%s) Database %s already does not exist.", h.IP, dbname, err))
		return
	}

	// TODO: How do we drop a database in bdr properly?
	sq := fmt.Sprintf(`DROP DATABASE IF EXISTS %s`, dbname)
	log.Trace(fmt.Sprintf(`RDPG#DropDatabase(%s) %s DROP > %s`, dbname, h.IP, sq))
	_, err = db.Exec(sq)
	if err != nil {
		log.Error(fmt.Sprintf("RDPG#DropDatabase(%s) DROP %s ! %s", dbname, h.IP, err))
		return
	}
	return
}

func (h *Host) PGDropUser(dbuser string) (err error) {
	h.Set(`database`, `postgres`)
	db, err := h.PGConnect()
	if err != nil {
		log.Error(fmt.Sprintf("pg.Host<%s>#PGDropUser(%s) %s ! %s", h.IP, dbuser, h.URI, err))
		return
	}
	defer db.Close()

	exists, err := h.PGUserExists(dbuser)
	if err != nil {
		log.Error(fmt.Sprintf("pg.Host<%s>#PGDropUser(%s) ! %s", h.IP, dbuser, err))
		return
	}
	if !exists {
		log.Error(fmt.Sprintf("pg.Host<%s>#PGDropUser(%s) User %s already does not exist.", h.IP, dbuser, err))
		return
	}

	// TODO: How do we drop a database in bdr properly?
	sq := fmt.Sprintf(`DROP USER %s`, dbuser)
	log.Trace(fmt.Sprintf(`RDPG#DropDatabase(%s) %s DROP > %s`, dbuser, h.IP, sq))
	_, err = db.Exec(sq)
	if err != nil {
		log.Error(fmt.Sprintf("RDPG#DropDatabase(%s) DROP %s ! %s", dbuser, h.IP, err))
		return
	}

	return
}

// Set host property to given value then regenerate the URI and DSN properties.
func (h *Host) Set(key, value string) (err error) {
	switch key {
	case "ip":
		h.IP = value
	case "port":
		h.Port = value
	case "user":
		h.User = value
	case "database":
		h.Database = value
	case "connect_timeout":
		h.ConnectTimeout = value
	case "sslmode":
		h.SSLMode = value
	case "pass":
	case "default": // A Bug
		err = fmt.Errorf(`Attempt to set unknown key %s to value %s for host %+v.`, key, value, *h)
		return err
	}
	h.pgURI()
	h.pgDSN()

	return
}

// Build and set the host's URI property
func (h *Host) pgURI() {
	d := `postgres://%s@%s:%s/%s?fallback_application_name=%s&connect_timeout=%s&sslmode=%s`
	h.URI = fmt.Sprintf(d, h.User, h.IP, h.Port, h.Database, `rdpg`, `5`, `disable`)
	return
}

// Build and set the host's DSN property
func (h *Host) pgDSN() {
	d := `user=%s host=%s port=%s dbname=%s fallback_application_name=%s connect_timeout=%s sslmode=%s`
	h.DSN = fmt.Sprintf(d, h.User, h.IP, h.Port, h.Database, `rdpg`, `5`, `disable`)
	return
}

// Connect to the host's database and return database connection object if successful
func (h *Host) PGConnect() (db *sqlx.DB, err error) {
	db, err = sqlx.Connect(`postgres`, h.URI)
	if err != nil {
		log.Error(fmt.Sprintf(`pg.Host<%s>#Connect() %s ! %s`, h.IP, h.URI, err))
		return db, err
	}
	return db, nil
}
