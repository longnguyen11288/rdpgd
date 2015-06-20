package rdpg

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"syscall"

	"github.com/jmoiron/sqlx"
	"github.com/wayneeseguin/rdpg-agent/log"
)

var (
	rdpgURI string
)

type RDPG struct {
	URI string
	DB  *sqlx.DB
}

func init() {
	rdpgURI = os.Getenv("RDPG_ADMIN_PG_URI")
	if rdpgURI == "" || rdpgURI[0:13] != "postgresql://" {
		log.Error("ERROR: RDPG_ADMIN_PG_URI is not set.")
		proc, _ := os.FindProcess(os.Getpid())
		proc.Signal(syscall.SIGTERM)
	}

	db, err := sqlx.Connect("postgres", rdpgURI)
	if err != nil {
		log.Error(fmt.Sprintf("Failed connecting to %s err: %s", rdpgURI, err))
		proc, _ := os.FindProcess(os.Getpid())
		proc.Signal(syscall.SIGTERM)
	}
	err = db.Ping()
	if err != nil {
		db.Close()
		log.Error(fmt.Sprintf("Unable to Ping %s err: %s", rdpgURI, err))
		proc, _ := os.FindProcess(os.Getpid())
		proc.Signal(syscall.SIGTERM)
	}
	db.Close()
}

// TODO: RDPG Struct => RDPG Struct, allowing for multiple instances of RDPG
// TODO: Add concept of 'target' RDPG Cluster instead of assuming local.
func NewRDPG(uri string) *RDPG {
	if uri == "" || uri[0:13] != "postgresql://" {
		log.Error(fmt.Sprintf("rdpg.NewRDPG() uri malformed ! %s", uri))
		return nil
	}
	if uri != "" {
		return &RDPG{URI: uri}
	} else {
		return &RDPG{URI: rdpgURI}
	}
}

func (r *RDPG) connect() (db *sqlx.DB, err error) {
	db, err = sqlx.Connect("postgres", r.URI)
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.Host#Connect() %s ! %s", r.URI, err))
	}
	return db, err
}

// TODO: Instead pass back *sql.DB
func (r *RDPG) OpenDB(dbname string) error {
	if r.DB == nil {
		u, err := url.Parse(r.URI)
		if err != nil {
			log.Error(fmt.Sprintf("Failed parsing URI %s err: %s", r.URI, err))
		}
		u.Path = dbname
		r.URI = u.String()
		db, err := sqlx.Connect("postgres", r.URI)
		if err != nil {
			log.Error(fmt.Sprintf("Failed connecting to %s err: %s", rdpgURI, err))
			return err
		}
		r.DB = db
	} else {
		err := r.DB.Ping()
		if err != nil {
			db, err := sqlx.Connect("postgres", r.URI)
			if err != nil {
				log.Error(fmt.Sprintf("Failed connecting to %s err: %s", rdpgURI, err))
				proc, _ := os.FindProcess(os.Getpid())
				proc.Signal(syscall.SIGTERM)
				return err
			}
			r.DB = db
		}
	}
	return nil
}

func (r *RDPG) Hosts() (hosts []Host) {
	db, err := r.connect()
	if err != nil {
		log.Error(fmt.Sprintf("RDPG#Hosts() ! %s", err))
	}

	// TODO: Populate list of rdpg hosts for given URL,
	//`SELECT node_local_dsn FROM bdr.bdr_nodes INTO rdpg.hosts (node_local_dsn);`

	type dsn struct {
		DSN string `db:"node_local_dsn"`
	}

	dsns := []dsn{}
	err = db.Select(&dsns, SQL["bdr_nodes_dsn"])
	if err != nil {
		log.Error(fmt.Sprintf("RDPG#Hosts() %s ! %s", SQL["bdr_nodes"], err))
	}

	for _, t := range dsns {
		host := Host{}
		s := strings.Split(t.DSN, " ")
		host.LocalDSN = t.DSN
		host.Host = strings.Split(s[0], "=")[1]
		host.Port = strings.Split(s[1], "=")[1]
		host.User = strings.Split(s[2], "=")[1]
		host.Database = `postgres` // strings.Split(s[3], "=")[1]
		hosts = append(hosts, host)
	}
	// TODO: Get this information into the database and then out of the rdpg.hosts
	//rows, err := db.Query("SELECT host,port,user,'postgres' FROM rdpg.hosts;")
	//if err != nil {
	//	log.Error(fmt.Sprintf("Hosts() %s", err))
	//} else {
	//	sqlx.StructScan(rows, hosts)
	//}
	db.Close()
	return hosts
}
