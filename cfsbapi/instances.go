package cfsbapi

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/wayneeseguin/rdpgd/log"
	"github.com/wayneeseguin/rdpgd/rdpg"
)

type Instance struct {
	Id             string `db:"id"`
	ClusterID      string `db:"cluster_id" json:"cluster_id"`
	InstanceId     string `db:"instance_id" json:"instance_id"`
	ServiceId      string `db:"service_id" json:"service_id"`
	PlanId         string `db:"plan_id" json:"plan_id"`
	OrganizationId string `db:"organization_id" json:"organization_id"`
	SpaceId        string `db:"space_id" json:"space_id"`
	Database       string `db:"dbname" json:"dbname"`
	User           string `db:"dbuser" json:"uname"`
	Pass           string `db:"dbpass" json:"pass"`
	lock           *consulapi.Lock
	lockCh         <-chan struct{}
}

func NewServiceInstance(instanceId, serviceId, planId, organizationId, spaceId string) (i *Instance, err error) {
	re := regexp.MustCompile("[^A-Za-z0-9_]")
	id := instanceId
	identifier := strings.ToLower(string(re.ReplaceAll([]byte(id), []byte(""))))
	i = &Instance{
		InstanceId:     strings.ToLower(instanceId),
		ServiceId:      strings.ToLower(serviceId),
		PlanId:         strings.ToLower(planId),
		OrganizationId: strings.ToLower(organizationId),
		SpaceId:        strings.ToLower(spaceId),
		Database:       "d" + identifier,
		User:           "u" + identifier,
	}
	if i.ServiceId == "" {
		err = errors.New("Service ID is required.")
		return
	}
	if i.PlanId == "" {
		err = errors.New("Plan ID is required.")
		return
	}
	if i.OrganizationId == "" {
		err = errors.New("OrganizationId ID is required.")
		return
	}
	if i.SpaceId == "" {
		err = errors.New("Space ID is required.")
		return
	}
	return
}

func ActiveInstances() (is []Instance, err error) {
	r := rdpg.NewRDPG()
	sq := ` SELECT id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass 
	FROM cfsbapi.instances 
	WHERE effective_at IS NOT NULL AND decommissioned_at IS NULL
	LIMIT 1; `
	err = r.OpenDB("rdpg")
	if err != nil {
		log.Error(fmt.Sprintf("cfsbapi#ActiveInstances() ! %s", err))
	}
	defer r.DB.Close()

	err = r.DB.Select(&is, sq)
	if err != nil {
		// TODO: Change messaging if err is sql.NoRows then say couldn't find instance with instanceId
		log.Error(fmt.Sprintf("cfsbapi.ActiveInstances() ! %s", err))
	}
	return
}

func FindInstance(instanceId string) (i *Instance, err error) {
	r := rdpg.NewRDPG()
	in := Instance{}
	sq := `SELECT id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass FROM cfsbapi.instances WHERE instance_id=lower(?) LIMIT 1;`
	err = r.OpenDB("rdpg")
	if err != nil {
		log.Error(fmt.Sprintf("cfsbapi#CreateBinding() ! %s", err))
	}
	defer r.DB.Close()

	err = r.DB.Get(&in, sq, instanceId)
	if err != nil {
		// TODO: Change messaging if err is sql.NoRows then say couldn't find instance with instanceId
		log.Error(fmt.Sprintf("cfsbapi.FindInstance(%s) ! %s", instanceId, err))
	}
	i = &in
	return
}

func (i *Instance) Provision() (err error) {
	// TODO: fetch precreated instance
	// bind it to this instanceId
	r := rdpg.NewRDPG()
	err = r.OpenDB("rdpg")
	if err != nil {
		log.Error(fmt.Sprintf("cfsbapi#Provision(%s) ! %s", i.InstanceId, err))
	}
	defer r.DB.Close()

	var id int

	// TODO: Compute which cluster the database will be assigned to based on
	//      min(# assigned for each cluster), then targeting this cluster:
	// TODO: Take into account plan with the above calculation, eg. dedicated vs shared
	for { // In case we need to wait for a precreated database on the cluster...
		log.Trace(fmt.Sprintf(`Querying for a pre-created instance...`))
		sq := `SELECT id FROM cfsbapi.instances WHERE instance_id IS NULL LIMIT 1;`
		err = r.DB.Get(&id, sq)
		if err != nil {
			log.Error(fmt.Sprintf("cfsbapi.Instance#Provision(%s) ! %s", i.InstanceId, err))
			return
		}
		log.Trace(fmt.Sprintf(`cfsbapi.Instance#Provision(%s) > Attempting to lock instance %s.`, i.InstanceId, id))

		i.Id = string(id)

		err = i.Lock()
		if err != nil {
			log.Error(fmt.Sprintf("cfsbapi.Instance#Provision(%s) Failed Locking instance %d ! %s", i.InstanceId, id, err))
			time.Sleep(5 * time.Second) // Wait a second...
			continue                    // ...then try again
		}

		sq = `UPDATE cfsbapi.instances SET instance_id = ? service_id = ? plan_id = ? organization_id = ? space_id = ? WHERE id=?`
		_, err = r.DB.Exec(sq, i.InstanceId, i.ServiceId, i.PlanId, i.OrganizationId, i.SpaceId)
		if err != nil {
			log.Error(fmt.Sprintf(`cfsbapi.Instance#Provision(%s) ! %s`, i.InstanceId, err))
			return
		}

		err = i.Unlock()
		if err != nil {
			log.Error(fmt.Sprintf(`cfsbapi.Instance#Provision(%s) Unlocking ! %s`, i.InstanceId, err))
			return
		}
		// TODO: Trigger enqueueing of database creation on target cluster via AdminAPI.
		// TODO: Also have scheduler which enqueues if number precreated databases < 10
		break
	}
	return
}

func (i *Instance) Remove() (err error) {
	r := rdpg.NewRDPG()
	err = r.OpenDB("rdpg")
	if err != nil {
		log.Error(fmt.Sprintf("Instance#Remove(%s) ! %s", i.InstanceId, err))
	}
	defer r.DB.Close()

	_, err = r.DB.Exec(`UPDATE cfsbapi.instances SET ineffective_at = CURRENT_TIMESTAMP WHERE id=?`, i.Id)
	if err != nil {
		log.Error(fmt.Sprintf("Instance#Remove(%s) ! %s", i.InstanceId, err))
	}
	time.Sleep(1) // Wait for the update to propigate to the other hosts.

	cluster, err := rdpg.NewCluster(i.ClusterID)
	if err != nil {
		log.Error(fmt.Sprintf(`cfsbapi.Instance#ExternalDNS(%s) ! %s`, i.InstanceId, err))
	}
	for _, node := range cluster.Nodes {
		err := rdpg.CallAdminAPI(node.PG.IP, "PUT", "services/pgbouncer/configure")
		if err != nil {
			log.Error(fmt.Sprintf(`Instance<%s>#Provision(%s) ! %s`, node.PG.IP, i.InstanceId, err))
		}
	}
	return
}

func (i *Instance) ExternalDNS() (dns string) {
	r := rdpg.NewRDPG()
	cluster, err := rdpg.NewCluster(r.ClusterID)
	if err != nil {
		log.Error(fmt.Sprintf(`cfsbapi.Instance#ExternalDNS(%s) ! %s`, i.InstanceId, err))
	}
	node, err := cluster.WriteMaster()
	if err != nil {
		log.Error(fmt.Sprintf(`cfsbapi.Instance#ExternalDNS(%s) ! %s`, i.InstanceId, err))
	}

	// TODO: Figure out where we'll store and retrieve the external DNS information instead of IP

	return node.PG.IP + ":5432"
}

func (i *Instance) URI() (uri string) {
	d := `postgres://%s:%s@%s/%s?connect_timeout=%s&sslmode=%s`
	uri = fmt.Sprintf(d, i.User, i.Pass, i.ExternalDNS(), i.Database, `5`, `disable`)
	return
}

func (i *Instance) DSN() (uri string) {
	dns := i.ExternalDNS()
	s := strings.Split(dns, ":")
	d := `host=%s port=%s user=%s password=%s dbname=%s connect_timeout=%s sslmode=%s`
	uri = fmt.Sprintf(d, s[0], s[1], i.User, i.Pass, i.Database, `5`, `disable`)
	return
}

func (i *Instance) JDBCURI() (uri string) {
	dns := i.ExternalDNS()
	s := strings.Split(dns, ":")
	d := `host=%s port=%s user=%s password=%s dbname=%s connect_timeout=%s sslmode=%s`
	uri = fmt.Sprintf(d, s[0], s[1], i.User, i.Pass, i.Database, `5`, `disable`)
	return
}

func Instances() (si []Instance, err error) {
	r := rdpg.NewRDPG()
	err = r.OpenDB("rdpg")
	if err != nil {
		log.Error(fmt.Sprintf("cfsbapi#Instances() ! %s", err))
	}
	defer r.DB.Close()

	si = []Instance{}
	// TODO: Move this into a versioned SQL Function.
	sq := `SELECT instance_id, service_id, plan_id, organization_id, space_id, dbname, dubser, 'md5'||md5(cfsbapi.instances.dbpass||dubser) as dbpass FROM cfsbapi.instances WHERE ineffective_at IS NULL; `
	err = r.DB.Select(&si, sq)
	if err != nil {
		// TODO: Change messaging if err is sql.NoRows then say couldn't find instance with instanceId
		log.Error(fmt.Sprintf("cfsbapi.Instances() ! %s", err))
	}
	return
}

func Remove() (err error) {
	return
}

func (i *Instance) Lock() (err error) {
	// Acquire consul schedulerLock to aquire right to schedule tasks.
	r := rdpg.NewRDPG()
	key := fmt.Sprintf("rdpg/%s/cfsb/instance/id/%s", r.ClusterID, i.Id)
	client, _ := consulapi.NewClient(consulapi.DefaultConfig())
	i.lock, err = client.LockKey(key)
	if err != nil {
		log.Error(fmt.Sprintf("scheduler.Schedule() Error Locking Scheduler Key %s ! %s", key, err))
		return
	}

	i.lockCh, err = i.lock.Lock(nil) // Acquire Consul K/V Lock
	if err != nil {
		log.Error(fmt.Sprintf("scheduler.Lock() Error Aquiring Scheduler Key lock %s ! %s", key, err))
		return
	}

	if i.lockCh == nil {
		err = fmt.Errorf(`Scheduler Lock not aquired.`)
	}
	return
}

func (i *Instance) Unlock() (err error) {
	if i.lock != nil {
		err = i.lock.Unlock()
	}
	return
}
