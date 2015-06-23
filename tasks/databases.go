package tasks

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/wayneeseguin/rdpgd/bdr"
	"github.com/wayneeseguin/rdpgd/cfsbapi"
	"github.com/wayneeseguin/rdpgd/log"
	"github.com/wayneeseguin/rdpgd/rdpg"
	"github.com/wayneeseguin/rdpgd/uuid"
)

// Pre-Create Database as a background task in advance of being requested.
func PrecreateDatabase(data string) (err error) {
	// Data is the cluster id (consul datacenter)
	// For now we assume data is simply the database name.
	dc := data
	b := bdr.NewBDR(dc)

	re := regexp.MustCompile("[^A-Za-z0-9_]")

	u1 := uuid.NewUUID().String()
	u2 := uuid.NewUUID().String()
	identifier := strings.ToLower(string(re.ReplaceAll([]byte(u1), []byte(""))))
	dbpass := strings.ToLower(string(re.ReplaceAll([]byte(u2), []byte(""))))

	i := &cfsbapi.Instance{
		Database: "d" + identifier,
		User:     "u" + identifier,
		Pass:     dbpass,
	}

	// TODO: Keep the databases under rdpg schema, link to them in the
	// cfsb.instances table so that we separate the concerns of CF and databases.

	err = b.CreateUser(i.User, i.Pass)
	if err != nil {
		log.Error(fmt.Sprintf("tasks.CreateDatabase(%s) CreateUser(%s) ! %s", i.InstanceId, i.User, err))
		return err
	}

	err = b.CreateDatabase(i.Database, i.User)
	if err != nil {
		log.Error(fmt.Sprintf("tasks.CreateDatabase(%s) CreateDatabase(%s,%s) ! %s", i.InstanceId, i.Database, i.User, err))
		return err
	}

	err = b.CreateReplicationGroup(i.Database)
	if err != nil {
		log.Error(fmt.Sprintf("tasks.CreateDatabase(%s) CreateReplicationGroup(%s) ! %s", i.InstanceId, i.Database, err))
		return err
	}

	r := rdpg.NewRDPG()
	err = r.OpenDB("rdpg")
	if err != nil {
		log.Error(fmt.Sprintf(`tasks.CreateDatabase(%s) ! %s`, i.InstanceId, err))
	}
	defer r.DB.Close()

	sq := `INSERT INTO cfsbapi.instances (dbname, dbuser, dbpass) VALUES (?,?,?);`
	log.Trace(fmt.Sprintf(`tasks.CreateDatabase(%s) > %s`, i.InstanceId, sq))
	_, err = r.DB.Query(sq, i.Database, i.User, i.Pass)
	if err != nil {
		log.Error(fmt.Sprintf(`tasks.CreateDatabase(%s) ! %s`, i.InstanceId, err))
	}
	// TODO: Insert task to create new database.
	return
}

// TODO: This should be remove database
func RemoveDatabase(data string) (err error) {
	// For now we assume data is simply the database name.
	r := rdpg.NewRDPG()
	key := fmt.Sprintf("rdpg/%s/work/databases/remove", r.ClusterID)
	client, _ := api.NewClient(api.DefaultConfig())
	lock, err := client.LockKey(key)
	if err != nil {
		log.Error(fmt.Sprintf("worker.RemoveDatabase() Error aquiring lock ! %s", err))
		return
	}
	leaderCh, err := lock.Lock(nil)
	if err != nil {
		log.Error(fmt.Sprintf("worker.RemoveDatabase() Error aquiring lock ! %s", err))
		return
	}
	if leaderCh == nil {
		log.Trace(fmt.Sprintf("worker.RemoveDatabase() > Not Leader."))
		return
	}
	log.Trace(fmt.Sprintf("worker.RemoveDatabase() > Leader."))

	ids := []string{}
	sq := fmt.Sprintf(`SELECT instance_id from cfsbapi.instances WHERE ineffective_at IS NOT NULL AND ineffective_at < CURRENT_TIMESTAMP AND decommissioned_at IS NULL`)
	err = r.DB.Select(&ids, sq)
	if err != nil {
		log.Error(fmt.Sprintf("worker.RemoveDatabase() Querying for Databases to Cleanup ! %s", err))
	}

	for _, id := range ids {

		// TODO: Find cluster datacenter that instance is on.... then create NEWBDR
		// for that cluster:
		uri := "postgres://"
		b := bdr.NewBDR(uri)

		i, err := cfsbapi.FindInstance(id)
		if err != nil {
			log.Error(fmt.Sprintf("worker.RemoveDatabase(%s) FindingInstance(%s) ! %s", i.Database, i.InstanceId, err))
			r.DB.Close()
			continue
		}

		err = b.DisableDatabase(i.Database)
		if err != nil {
			log.Error(fmt.Sprintf("worker.RemoveDatabase() DisableDatabase(%s) for %s ! %s", i.Database, i.InstanceId, err))
			r.DB.Close()
			continue
		}

		err = b.BackupDatabase(i.Database)
		if err != nil {
			log.Error(fmt.Sprintf("worker.RemoveDatabase() BackupDatabase(%s) ! %s", i.Database, err))
			r.DB.Close()
			continue
		}

		// Question, do we need to "stop" the replication group before dropping the database?
		err = b.DropDatabase(i.Database)
		if err != nil {
			log.Error(fmt.Sprintf("worker.RemoveDatabase() DropDatabase(%s) for %s ! %s", i.Database, i.InstanceId, err))
			r.DB.Close()
			continue
		}

		err = b.DropUser(i.User)
		if err != nil {
			log.Error(fmt.Sprintf("worker.RemoveDatabase() DropUser(%s) for %s ! %s", i.User, i.InstanceId, err))
			r.DB.Close()
			continue
		}

		err = b.DropDatabase(i.Database)
		if err != nil {
			log.Error(fmt.Sprintf("worker.RemoveDatabase() DropDatabase(%s) for %s ! %s", i.Database, i.InstanceId, err))
			r.DB.Close()
			continue
		}
	}
	r.DB.Close()

	return
}
