package rdpg

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"syscall"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/jmoiron/sqlx"
	"github.com/wayneeseguin/rdpgd/bdr"
	"github.com/wayneeseguin/rdpgd/log"
	"github.com/wayneeseguin/rdpgd/pg"
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

	r.ClusterID = info["Config"]["Datacenter"].(string)
	r.IP = info["Config"]["AdvertiseAddr"].(string)

	return
}

func (r *RDPG) SetURI(uri string) (err error) {
	// TODO: SetURI likely is not needed anymore.
	if uri == "" || uri[0:13] != "postgresql://" {
		// TODO: use uri.Parse to further validate URI.
		err = fmt.Errorf(`Malformed postgresql:// URI`, uri)
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

// Bootstrap the RDPG Database and associated services.
func (r *RDPG) Bootstrap(role string) (err error) {
	port := os.Getenv("PG_PORT")
	if port == "" {
		port = "5432"
	}
	p := pg.NewPG(`127.0.0.1`, port, `postgres`, `postgres`)

	os.Getenv("RDPG")

	exists, err := p.UserExists(`rdpg`)
	if err != nil {
	}
	if !exists {
		pass := os.Getenv(`RDPGD_PG_PASS`)
		p.CreateUser(`rdpg`, pass)
	}

	db, err := p.Connect()
	if err != nil {
		log.Error(fmt.Sprintf("r.RDPG#Bootstrap() ! %s", err))
		return
	}
	defer db.Close()

	sq := `ALTER USER rdpg WITH SUPERUSER CREATEDB CREATEROLE INHERIT;`
	_, err = db.Exec(sq)
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.Bootstrap() ! %s", err))
		return
	}

	exists, err = p.DatabaseExists(`rdpg`)
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.Bootstrap() ! %s", err))
		return
	}
	if !exists {
		sq := `CREATE DATABASE rdpg WITH OWNER rdpg TEMPLATE template0 ENCODING 'UTF8'`
		_, err = db.Query(sq)
		if err != nil {
			log.Error(fmt.Sprintf("rdotstrap() ! %s", err))
			return
		}
	}

	err = p.CreateExtensions(`rdpg`, []string{"btree_gist", "bdr"})
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.Bootstrap() ! %s", err))
		return
	}

	// Node is ready, now we bootstrap the cluster if we get the lock.

	key := fmt.Sprintf("rdpg/%s/bootstrap", r.ClusterID)
	client, _ := consulapi.NewClient(consulapi.DefaultConfig())
	lock, err := client.LockKey(key)
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.Bootstrap() Error Locking Bootstrap Key %s ! %s", key, err))
		return
	}
	lockCh, err := lock.Lock(nil) // Acquire Consul K/V Lock
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.Bootstrap() Error Aquiring Bootstrap Key lock %s ! %s", key, err))
		return
	}
	if lockCh == nil {
		err = fmt.Errorf(`rdpg.Bootstrap() Bootstrap Lock not aquired.`)
	}
	// Note that we are intentionally not unlocking the cluster bootsrap if
	// bootstrapping was successful as we only want it done once per cluster.

	b := bdr.NewBDR(r.ClusterID)
	err = b.CreateReplicationGroup(`rdpg`)
	if err != nil {
		lock.Unlock()
		log.Error(fmt.Sprintf(`rdpg.Bootstrap() Error Starting Replication for database 'rdpg' ! %s`, err))
		return
	}

	_, err = r.DB.Exec(`SELECT bdr.bdr_node_join_wait_for_ready();`)
	if err != nil {
		log.Error(fmt.Sprintf("RDPG#initSchema() bdr.bdr_node_join_wait_for_ready ! %s", err))
		return
	}

	err = r.InitSchema()
	if err != nil {
		err = lock.Unlock()
		if err != nil {
			log.Error(fmt.Sprintf("rdpg.Bootstrap() Error Unlocking Scheduler ! %s", err))
			return
		}
	}
	cluster, err := NewCluster(r.ClusterID)
	if err != nil {
		log.Error(fmt.Sprintf(`rdpg.Bootstrap() ! %s`, err))
		return
	}
	for index, node := range cluster.Nodes {
		// Set write master if it is the first node
		if index == 0 {
			err = cluster.SetWriteMaster(&node)
			if err != nil {
				log.Error(fmt.Sprintf(`rdpg.Bootstrap() Setting Write Master ! %s`, err))
				return
			}
		}
		// Reconfigure PGBouncer
		err = CallAdminAPI(node.PG.IP, "PUT", "services/pgbouncer/configure")
		if err != nil {
			log.Error(fmt.Sprintf(`rdpg.Bootstrap(%s) reconfigure pgbouncer %s ! %s`, node.PG.IP, err))
			return
		}
		// Reconfigure HAProxy
		err = CallAdminAPI(node.PG.IP, "PUT", "services/haproxy/configure")
		if err != nil {
			log.Error(fmt.Sprintf(`rdpg.Bootstrap(%s) reconfigure haroxy %s ! %s`, node.PG.IP, err))
			return
		}
	}
	return
}
