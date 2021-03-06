package rdpg

import (
	"fmt"
	"net/url"
	"os"
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

func New() *RDPG {
	return &RDPG{URI: rdpgURI}
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
