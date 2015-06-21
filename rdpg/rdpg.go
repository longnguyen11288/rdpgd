package rdpg

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"syscall"

	"github.com/jmoiron/sqlx"
	"github.com/wayneeseguin/rdpgd/log"
)

var (
	rdpgURI string
)

type RDPG struct {
	URI string
	DB  *sqlx.DB
}

func init() {
	// Question: should we not bother trying to connect until first use???
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
func NewRDPG() (r *RDPG) {
	r = &RDPG{URI: rdpgURI}
	return
}

func (r *RDPG) SetURI(uri string) (err error) {
	if uri == "" || uri[0:13] != "postgresql://" {
		// TODO: use uri.Parse to further validate URI.
		err = fmt.Errorf(`Malformed postgresql:// URI`, uri)
		log.Error(fmt.Sprintf("rdpg.NewRDPG() uri malformed ! %s", err))
		return
	}
	r.URI = uri
	return
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

func (r *RDPG) connect() (db *sqlx.DB, err error) {
	db, err = sqlx.Connect("postgres", r.URI)
	if err != nil {
		log.Error(fmt.Sprintf("rdpg#connect() %s ! %s", r.URI, err))
	}
	return db, err
}

// Call RDPG Admin API for given IP
func CallAdminAPI(ip, method, path string) (err error) {
	url := fmt.Sprintf("http://%s:%s/%s", ip, os.Getenv("RDPG_ADMIN_PORT"), path)
	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(`{}`)))
	// req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(os.Getenv("RDPG_ADMIN_USER"), os.Getenv("RDPG_ADMIN_PASS"))
	client := &http.Client{}

	log.Trace(fmt.Sprintf(`pg.Host<%s>#AdminAPI(%s,%s) %s`, ip, method, path, url))
	resp, err := client.Do(req)
	if err != nil {
		log.Error(fmt.Sprintf(`pg.Host<%s>#AdminAPI(%s,%s) ! %s`, ip, method, url, err))
	}
	resp.Body.Close()
	return
}
