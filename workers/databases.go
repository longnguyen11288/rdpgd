package workers

import (
	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/wayneeseguin/rdpg-agent/cfsbapi"
	"github.com/wayneeseguin/rdpg-agent/log"
	"github.com/wayneeseguin/rdpg-agent/rdpg"
)

// TODO: This should be remove database
func RemoveDatabase(data string) (err error) {
	client, _ := api.NewClient(api.DefaultConfig())
	lock, err := client.LockKey("rdpg/work/databases/remove/lock")
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

	r := rdpg.New()
	ids := []string{}
	sq := fmt.Sprintf(`SELECT instance_id from cfsbapi.instances WHERE ineffective_at IS NOT NULL AND ineffective_at < CURRENT_TIMESTAMP AND decommissioned_at IS NULL`)
	err = r.DB.Select(&ids, sq)
	if err != nil {
		log.Error(fmt.Sprintf("worker.RemoveDatabase() Querying for Databases to Cleanup ! %s", err))
	}

	for _, id := range ids {
		i, err := cfsbapi.FindInstance(id)
		if err != nil {
			log.Error(fmt.Sprintf("worker.RemoveDatabase(%s) FindingInstance(%s) ! %s", i.Database, i.InstanceId, err))
			r.DB.Close()
			continue
		}

		err = r.DisableDatabase(i.Database)
		if err != nil {
			log.Error(fmt.Sprintf("worker.RemoveDatabase() DisableDatabase(%s) for %s ! %s", i.Database, i.InstanceId, err))
			r.DB.Close()
			continue
		}

		err = r.BackupDatabase(i.Database)
		if err != nil {
			log.Error(fmt.Sprintf("worker.RemoveDatabase() BackupDatabase(%s) ! %s", i.Database, err))
			r.DB.Close()
			continue
		}

		// Question, do we need to "stop" the replication group before dropping the database?
		err = r.DropDatabase(i.Database)
		if err != nil {
			log.Error(fmt.Sprintf("worker.RemoveDatabase() DropDatabase(%s) for %s ! %s", i.Database, i.InstanceId, err))
			r.DB.Close()
			continue
		}

		err = r.DropUser(i.User)
		if err != nil {
			log.Error(fmt.Sprintf("worker.RemoveDatabase() DropUser(%s) for %s ! %s", i.User, i.InstanceId, err))
			r.DB.Close()
			continue
		}

		err = r.DropDatabase(i.Database)
		if err != nil {
			log.Error(fmt.Sprintf("worker.RemoveDatabase() DropDatabase(%s) for %s ! %s", i.Database, i.InstanceId, err))
			r.DB.Close()
			continue
		}
	}
	r.DB.Close()

	return
}
