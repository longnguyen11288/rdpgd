package rdpg

import (
	"fmt"
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
	defer db.Close()

	err = db.Ping()
	if err != nil {
		db.Close()
		log.Error(fmt.Sprintf("Unable to Ping %s err: %s", rdpgURI, err))
		proc, _ := os.FindProcess(os.Getpid())
		proc.Signal(syscall.SIGTERM)
	}

	initSchema(db)
}

func New() *RDPG {
	return &RDPG{URI: rdpgURI}
}

func (r *RDPG) OpenDB() error {
	if r.DB == nil {
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
