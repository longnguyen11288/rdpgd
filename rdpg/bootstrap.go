package rdpg

import (
	"fmt"
	"os"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/wayneeseguin/rdpgd/bdr"
	"github.com/wayneeseguin/rdpgd/log"
	"github.com/wayneeseguin/rdpgd/pg"
)

// Bootstrap the RDPG Database and associated services.
func (r *RDPG) Bootstrap(role string) (err error) {
	log.Info(`Bootstrapping...`)
	port := os.Getenv("PG_PORT")
	if port == "" {
		port = "5432"
	}
	p := pg.NewPG(`127.0.0.1`, port, `postgres`, `postgres`)
	exists, err := p.UserExists(`rdpg`)
	if err != nil {
		log.Error(fmt.Sprintf("r.RDPG#Bootstrap() UserExists() ! %s", err))
		return
	}
	if !exists {
		pass := os.Getenv(`RDPGD_PG_PASS`)
		err = p.CreateUser(`rdpg`, pass)
		if err != nil {
			log.Error(fmt.Sprintf("r.RDPG#Bootstrap() CreateUser() ! %s", err))
			return
		}
	}
	db, err := p.Connect()
	if err != nil {
		log.Error(fmt.Sprintf("r.RDPG#Bootstrap() Connect() ! %s", err))
		return
	}
	defer db.Close()

	sq := `ALTER USER rdpg WITH SUPERUSER CREATEDB CREATEROLE INHERIT;`
	_, err = db.Exec(sq)
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.Bootstrap() ALTER USER ! %s", err))
		return
	}
	exists, err = p.DatabaseExists(`rdpg`)
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.Bootstrap() DatabaseExists() ! %s", err))
		return
	}
	if !exists {
		err = p.CreateDatabase(`rdpg`, `rdpg`)
		if err != nil {
			log.Error(fmt.Sprintf("rdpg.Bootstrapping() CreateDatabase() ! %s", err))
			return
		}
	}
	err = p.CreateExtensions(`rdpg`, []string{"btree_gist", "bdr"})
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.Bootstrap() CreateExtensions() ! %s", err))
		return
	}

	// Node is ready, now we bootstrap the cluster if we get the lock.
	key := fmt.Sprintf("rdpg/%s/bootstrap", r.ClusterID)
	client, _ := consulapi.NewClient(consulapi.DefaultConfig())
	lock, err := client.LockKey(key)
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.Bootstrap() LockKey() Error Locking Bootstrap Key %s ! %s", key, err))
		return
	}
	lockCh, err := lock.Lock(nil) // Acquire Consul K/V Lock
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.Bootstrap() Lock() Error Aquiring Bootstrap Key lock %s ! %s", key, err))
		return
	}
	if lockCh == nil {
		log.Error(fmt.Sprintf("rdpg.Bootstrap() Bootstrap Lock not aquired, halting bootstrap."))
		return
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
		log.Error(fmt.Sprintf(`rdpg.Bootstrap() NewCluster() ! %s`, err))
		return
	}
	for index, node := range cluster.Nodes {
		// Set write master if it is the first node
		if index == 0 {
			err = cluster.SetWriteMaster(&node)
			if err != nil {
				log.Error(fmt.Sprintf(`rdpg.Bootstrap() SetWriteMaster() ! %s`, err))
				return
			}
		}
		// Reconfigure PGBouncer
		err = CallAdminAPI(node.PG.IP, "PUT", "services/pgbouncer/configure")
		if err != nil {
			log.Error(fmt.Sprintf(`rdpg.Bootstrap(%s) CallAdminAPI(pgbouncer/configure) %s ! %s`, node.PG.IP, err))
			return
		}
		// Reconfigure HAProxy
		err = CallAdminAPI(node.PG.IP, "PUT", "services/haproxy/configure")
		if err != nil {
			log.Error(fmt.Sprintf(`rdpg.Bootstrap(%s) CallAdminAPI(haroxy/configure) %s ! %s`, node.PG.IP, err))
			return
		}
	}
	return
}
