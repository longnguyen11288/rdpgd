package rdpg

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"syscall"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/jmoiron/sqlx"
	"github.com/wayneeseguin/rdpgd/log"
)

var (
	rdpgURI string
)

type RDPG struct {
	ClusterID string
	IP        string
	URI       string
	DB        *sqlx.DB
}

func init() {
	// Question: should we not bother trying to connect until first use???
	rdpgURI = os.Getenv("RDPGD_ADMIN_PG_URI")
	if rdpgURI == "" || rdpgURI[0:13] != "postgresql://" {
		log.Error("ERROR: RDPGD_ADMIN_PG_URI is not a proper postgresql:// URI.")
		proc, _ := os.FindProcess(os.Getpid())
		proc.Signal(syscall.SIGTERM)
	}
}

func NewRDPG() (r *RDPG) {
	r = &RDPG{URI: rdpgURI}

	client, err := consulapi.NewClient(consulapi.DefaultConfig())
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.NewRDPG() ! %s", err))
		return
	}
	agent := client.Agent()
	info, err := agent.Self()
	r.ClusterID = os.Getenv("RDPGD_CLUSTER")
	r.IP = info["Config"]["AdvertiseAddr"].(string)
	return
}

func (r *RDPG) SetURI(uri string) (err error) {
	// TODO: SetURI likely is not needed anymore.
	if uri == "" || uri[0:13] != "postgresql://" {
		// TODO: use uri.Parse to further validate URI.
		err = fmt.Errorf(`Malformed postgresql:// URI : %s`, uri)
		log.Error(fmt.Sprintf("rdpg.NewRDPG() uri malformed ! %s", err))
		return
	}
	r.URI = uri
	return
}

func Clusters() (clusters []string, err error) {
	client, err := consulapi.NewClient(consulapi.DefaultConfig())
	if err != nil {
		return
	}
	catalog := client.Catalog()
	clusters, err = catalog.Datacenters()
	if err != nil {
		return
	}
	return
}

// TODO: Instead pass back *sql.DB
func (r *RDPG) OpenDB(dbname string) (err error) {
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
		log.Error(fmt.Sprintf("rdpg.connect() Failed connecting to %s err: %s", r.URI, err))
	}
	return
}

// Call RDPG Admin API for given IP
func CallAdminAPI(ip, method, path string) (err error) {
	url := fmt.Sprintf("http://%s:%s/%s", ip, os.Getenv("RDPGD_ADMIN_PORT"), path)
	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(`{}`)))
	// req.Header.Set("Content-Type", "application/json")
	// TODO: Retrieve from configuration in database.
	req.SetBasicAuth(os.Getenv("RDPGD_ADMIN_USER"), os.Getenv("RDPGD_ADMIN_PASS"))

	client := &http.Client{}

	log.Trace(fmt.Sprintf(`pg.Host<%s>#AdminAPI(%s,%s) %s`, ip, method, path, url))
	resp, err := client.Do(req)
	if err != nil {
		log.Error(fmt.Sprintf(`pg.Host<%s>#AdminAPI(%s,%s) ! %s`, ip, method, url, err))
	}
	resp.Body.Close()
	return
}

func (r *RDPG) Register() (err error) {
	client, err := consulapi.NewClient(consulapi.DefaultConfig())
	if err != nil {
		log.Error(fmt.Sprintf("rdpgRDPG#Register() ! %s", err))
		return
	}
	agent := client.Agent()
	port, err := strconv.Atoi(os.Getenv("PG_PORT"))
	if err != nil {
		log.Error(fmt.Sprintf("rdpgRDPG#Register() ! %s", err))
		return
	}

	registration := &consulapi.AgentServiceRegistration{
		ID:   "rdpg",
		Name: r.ClusterID,
		Tags: []string{},
		Port: port,
		Check: &consulapi.AgentServiceCheck{
			HTTP:     fmt.Sprintf(`http://127.0.0.1:%s/health/pg`, os.Getenv("RDPGD_ADMIN_PORT")),
			Interval: "10s",
			Timeout:  "1s",
		},
	}
	agent.ServiceRegister(registration)
	return
}
