package rdpg

import (
	"fmt"
	"os"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/wayneeseguin/rdpgd/log"
	"github.com/wayneeseguin/rdpgd/pg"
)

var (
	bootstrapLeaderLock   *consulapi.Lock
	bootstrapLeaderLockCh <-chan struct{}
	bdrJoinLock           *consulapi.Lock
	bdrJoinLockCh         <-chan struct{}
	pgPort                string
	pgPass                string
	bdrLeaderIP           string
)

func init() {
	pgPort = os.Getenv("PG_PORT")
	if pgPort == "" {
		pgPort = "5432"
	}
	pgPass = os.Getenv(`RDPGD_PG_PASS`)
}

// Bootstrap the RDPG Database and associated services.
func (r *RDPG) Bootstrap(role string) (err error) {
	log.Info(`Bootstrapping...`)
	err = r.generalBootstrap()
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.Bootstrap() r.generalBootstrap() ! %s", err))
		return
	}
	leader, err := r.bootstrapLeaderLock()
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.Bootstrap() r.generalBootstrap() ! %s", err))
		return
	}
	if leader {
		r.leaderBootstrap()
	} else {
		r.nonLeaderBootstrap()
	}
	log.Trace(`Bootstrapping completed.`)
	return
}

// General Boostrapping that should occur on every node irrespective of role/leader.
func (r *RDPG) generalBootstrap() (err error) {
	p := pg.NewPG(`127.0.0.1`, pgPort, `postgres`, `postgres`, ``)

	exists, err := p.UserExists(`rdpg`)
	if err != nil {
		log.Error(fmt.Sprintf("r.RDPG#generalBootstrap() UserExists() ! %s", err))
		return
	}
	if !exists {
		err = p.CreateUser(`rdpg`, pgPass)
		if err != nil {
			log.Error(fmt.Sprintf("r.RDPG#generalBootstrap() CreateUser() ! %s", err))
			return
		}
	}
	db, err := p.Connect()
	if err != nil {
		log.Error(fmt.Sprintf("r.RDPG#generalBootstrap() Connect() ! %s", err))
		return
	}
	defer db.Close()
	sq := `ALTER USER rdpg WITH SUPERUSER CREATEDB CREATEROLE INHERIT;`
	_, err = db.Exec(sq)
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.generalBootstrap() ALTER USER ! %s", err))
		return
	}
	exists, err = p.DatabaseExists(`rdpg`)
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.generalBootstrap() DatabaseExists() ! %s", err))
		return
	}
	if !exists {
		err = p.CreateDatabase(`rdpg`, `rdpg`)
		if err != nil {
			log.Error(fmt.Sprintf("rdpg.generalBootstrapping() CreateDatabase() ! %s", err))
			return
		}
	}
	err = p.CreateExtensions(`rdpg`, []string{"btree_gist", "bdr"})
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.generalBootstrap() CreateExtensions() ! %s", err))
		return
	}
	err = r.Register()
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.generalBootstrap() Register() ! %s", err))
		return
	}
	return
}

// Attempt to obtain a lock on the boostrap leader.
func (r *RDPG) bootstrapLeaderLock() (locked bool, err error) {
	// Node is ready, now we bootstrap the cluster if we get the lock.
	client, _ := consulapi.NewClient(consulapi.DefaultConfig())
	key := fmt.Sprintf("rdpg/%s/bootstrap/leader/lock", r.ClusterID)
	bootstrapLeaderLock, err = client.LockKey(key)
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.bootstrapLeaderLock() LockKey() Error Locking Bootstrap Key %s ! %s", key, err))
		return
	}
	bootstrapLeaderLockCh, err := bootstrapLeaderLock.Lock(nil) // Acquire Consul K/V Lock
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.bootstrapLeaderLock() Lock() Error Aquiring Bootstrap Key lock %s ! %s", key, err))
		return
	}
	if bootstrapLeaderLockCh == nil {
		log.Error(fmt.Sprintf("rdpg.bootstrapLeaderLock() Bootstrap Lock not aquired, halting bootstrap."))
		return
	}
	return
}

// Unlock the bootstrap leader lock
func (r *RDPG) bootstrapLeaderUnlock() (err error) {
	err = bootstrapLeaderLock.Unlock()
	return
}

// Leader specific bootstrapping.
func (r *RDPG) leaderBootstrap() (err error) {
	client, _ := consulapi.NewClient(consulapi.DefaultConfig())
	agent := client.Agent()
	info, err := agent.Self()
	myIP := info["Config"]["AdvertiseAddr"].(string)
	kv := client.KV()
	key := fmt.Sprintf("rdpg/%s/bdr/join/ip", r.ClusterID)
	kvp, _, err := kv.Get(key, &consulapi.QueryOptions{})
	if err != nil {
		log.Error(fmt.Sprintf(`rdpg.leaderBootstrap() kv.Get() ! %s`, err))
		return
	}
	v := string(kvp.Value)
	if len(v) > 0 {
		// Skip, bdr group already created
	} else {
		// BDR Group not created yet, create and log myIP
		p := pg.NewPG(`127.0.0.1`, pgPort, `rdpg`, `rdpg`, pgPass)
		err = p.BDRGroupCreate(`rdpg`, `rdpg`)
		if err != nil {
			r.bootstrapLeaderUnlock()
			log.Error(fmt.Sprintf(`rdpg.Bootstrap() Error Starting Replication for database 'rdpg' ! %s`, err))
			return
		}
		kvp := &consulapi.KVPair{Key: key, Value: []byte(myIP)}
		_, err = kv.Put(kvp, &consulapi.WriteOptions{})
		if err != nil {
			log.Error(fmt.Sprintf(`rdpg.leaderBootstrap() ! %s`, err))
			return
		}
	}

	err = r.InitSchema()
	if err != nil {
		err = r.bootstrapLeaderUnlock()
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
		time.Sleep(1 * time.Second)
		// Trigger PGBouncer Reconfigure
		err = CallAdminAPI(node.PG.IP, "PUT", "services/pgbouncer/configure")
		if err != nil {
			log.Error(fmt.Sprintf(`rdpg.Bootstrap() CallAdminAPI(%s,pgbouncer/configure) ! %s`, node.PG.IP, err))
			return
		}
		// Trigger HAProxy Reconfigure
		err = CallAdminAPI(node.PG.IP, "PUT", "services/haproxy/configure")
		if err != nil {
			log.Error(fmt.Sprintf(`rdpg.Bootstrap() CallAdminAPI(%s,haroxy/configure) ! %s`, node.PG.IP, err))
			return
		}
	}
	return
}

// Non-Leader specifc bootstrapping.
func (r *RDPG) nonLeaderBootstrap() (err error) {
	client, _ := consulapi.NewClient(consulapi.DefaultConfig())
	kv := client.KV()
	key := fmt.Sprintf("rdpg/%s/bdr/join/ip", r.ClusterID)
	for { // for loop to get leader IP wait till ready
		kvp, _, err := kv.Get(key, &consulapi.QueryOptions{})
		if err != nil {
			log.Error(fmt.Sprintf(`rdpg#nonLeaderBootstrap() kv.Get() ! %s`, err))
			return err
		}
		v := string(kvp.Value)
		if len(v) == 0 {
			continue // Good to continue, bdr group was created...
		} else {
			bdrLeaderIP = v
			break
		}
	}

	err = r.bdrGroupJoin()
	if err != nil {
		log.Error(fmt.Sprintf(`rdpg#nonLeaderBootstrap() bdrGroupJoin() ! %s`, err))
	}

	err = r.reconfigureServices()
	if err != nil {
		log.Error(fmt.Sprintf(`rdpg#nonLeaderBootstrap() reconfigureServices() ! %s`, err))
	}

	return
}

// Attempt to obtain a lock on the boostrap leader.
func (r *RDPG) bdrJoinLock() (locked bool, err error) {
	// Node is ready, now we bootstrap the cluster if we get the lock.
	client, _ := consulapi.NewClient(consulapi.DefaultConfig())
	key := fmt.Sprintf("rdpg/%s/bdr/join/lock", r.ClusterID)
	bdrJoinLock, err = client.LockKey(key)
	if err != nil {
		log.Error(fmt.Sprintf("rdpg#bdrJoinLock() LockKey() Error Locking Bootstrap Key %s ! %s", key, err))
		return
	}
	bdrJoinLockCh, err := bdrJoinLock.Lock(nil) // Acquire Consul K/V Lock
	if err != nil {
		log.Error(fmt.Sprintf("rdpg#bdrJoinLock() Lock() Error Aquiring Bootstrap Key lock %s ! %s", key, err))
		return
	}
	if bdrJoinLockCh == nil {
		log.Error(fmt.Sprintf("rdpg#bdrJoinLock() Bootstrap Lock not aquired, halting bootstrap."))
		return
	}
	return
}

// Unlock the bootstrap leader lock
func (r *RDPG) bdrJoinUnlock() (err error) {
	err = bdrJoinLock.Unlock()
	if err != nil {
		log.Error(fmt.Sprintf(`rdpg#bdrJoinUnlock()! %s`, err))
	}
	return
}

func (r *RDPG) bdrGroupJoin() (err error) {
	for {
		locked, err := r.bdrJoinLock()
		if err != nil {
			log.Error(fmt.Sprintf(`rdpg#nonLeaderBootstrap() bdr join lock ! %s`, err))
		}
		if locked {
			log.Trace(fmt.Sprintf(`rdpg#nonLeaderBootstrap() Acquired bdr join lock`))
			break
		}
		time.Sleep(1 * time.Second)
	}

	p := pg.NewPG(`127.0.0.1`, pgPort, `rdpg`, `rdpg`, pgPass)
	joinPG := pg.NewPG(bdrLeaderIP, pgPort, `rdpg`, `rdpg`, pgPass)
	err = p.BDRGroupJoin(`rdpg`, `rdpg`, *joinPG)
	if err != nil {
		log.Error(fmt.Sprintf(`rdpg#nonLeaderBootstrap() Error Joining Replication for database 'rdpg' ! %s`, err))
		return
	}
	time.Sleep(3 * time.Second) // Wait 3 seconds for replication to replicate.
	return
}

func (r *RDPG) reconfigureServices() (err error) {
	// Trigger PGBouncer Reconfigure
	err = CallAdminAPI(`127.0.0.1`, "PUT", "services/pgbouncer/configure")
	if err != nil {
		log.Error(fmt.Sprintf(`rdpg#nonLeaderBootstrap() CallAdminAPI(127.0.0.1,pgbouncer/configure) ! %s`, err))
		return
	}

	// Trigger HAProxy Reconfigure
	err = CallAdminAPI(`127.0.0.1`, "PUT", "services/haproxy/configure")
	if err != nil {
		log.Error(fmt.Sprintf(`rdpg#nonLeaderBootstrap() CallAdminAPI(127.0.0.1,haroxy/configure) ! %s`, err))
		return
	}
	return
}
