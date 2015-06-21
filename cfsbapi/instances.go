package cfsbapi

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/armon/consul-api"
	"github.com/wayneeseguin/rdpgd/log"
	"github.com/wayneeseguin/rdpgd/rdpg"
)

type Instance struct {
	Id             string `db:"id"`
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
	r.OpenDB("rdpg")
	err = r.DB.Select(&is, sq)
	if err != nil {
		// TODO: Change messaging if err is sql.NoRows then say couldn't find instance with instanceId
		log.Error(fmt.Sprintf("cfsbapi.ActiveInstances() ! %s", err))
	}
	r.DB.Close()
	return
}

func FindInstance(instanceId string) (i *Instance, err error) {
	r := rdpg.NewRDPG()
	in := Instance{}
	sq := `SELECT id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass FROM cfsbapi.instances WHERE instance_id=lower($1) LIMIT 1;`
	r.OpenDB("rdpg")
	err = r.DB.Get(&in, sq, instanceId)
	if err != nil {
		// TODO: Change messaging if err is sql.NoRows then say couldn't find instance with instanceId
		log.Error(fmt.Sprintf("cfsbapi.FindInstance(%s) ! %s", instanceId, err))
	}
	r.DB.Close()
	i = &in
	return
}

func (i *Instance) Provision() (err error) {
	// TODO: fetch precreated instance
	// bind it to this instanceId
	r := rdpg.NewRDPG()
	r.OpenDB("rdpg")
	defer r.DB.Close()

	var id int

	// TODO: What if precreated queue is empty
	for { // In case we need to wait for one to be created...
		log.Trace(fmt.Trace(`Querying for a pre-created instance...`))
		sq := `SELECT id FROM cfsbapi.instances WHERE instance_id IS NULL LIMIT 1;`
		_, err = r.DB.Get(&id, sq)
		if err != nil {
			log.Error(fmt.Sprintf("cfsbapi.Instance#Provision(%s) ! %s", i.InstanceId, err))
			return
		}
		log.Trace(fmt.Trace(`cfsbapi.Instance#Provision(%s) > Attempting to lock instance %s.`, id))
		i.Id = id
		err = i.Lock()
		if err != nil {
			log.Error(fmt.Sprintf("cfsbapi.Instance#Provision(%s) Failed Locking instance %s ! %s", id, err))
			time.Sleep(1 * time.Milliseconds) // Wait a second...
			continue                          // ...then try again
		}
		sq = `UPDATE cfsbapi.instances SET instance_id = ? service_id = ? plan_id = ? organization_id = ? space_id = ? WHERE id=$1`
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
		// TODO: Enqueue pre-creation of another database.
		// TODO: Also have scheduler which enqueues if number precreated databases < 10
		break
	}
	return
}

func (i *Instance) Remove() (err error) {
	r := rdpg.NewRDPG()
	r.OpenDB("rdpg")
	_, err = r.DB.Exec(`UPDATE cfsbapi.instances SET ineffective_at = CURRENT_TIMESTAMP WHERE id=$1`, i.Id)
	if err != nil {
		log.Error(fmt.Sprintf("Instance#Remove(%s) ! %s", i.InstanceId, err))
	}

	time.Sleep(1) // Wait for the update to propigate to the other hosts.

	for _, host := range r.Hosts() {
		err := host.AdminAPI("PUT", "services/pgbouncer/configure")
		if err != nil {
			log.Error(fmt.Sprintf(`Instance#Provision(%s) %s ! %s`, i.InstanceId, host.IP, err))
		}
	}
	r.DB.Close()
	return
}

func (i *Instance) ExternalDNS() (dns string) {
	// TODO: Figure out where we'll store and retrieve the external DNS information
	r := rdpg.NewRDPG()
	hosts := r.Hosts()
	// TODO: Import the external DNS host via env variable configuration.
	return hosts[0].IP + ":5432"
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
	r.OpenDB("rdpg")
	si = []Instance{}
	// TODO: Move this into a versioned SQL Function.
	sq := `SELECT instance_id, service_id, plan_id, organization_id, space_id, dbname, dubser, 'md5'||md5(cfsbapi.instances.dbpass||dubser) as dbpass FROM cfsbapi.instances WHERE ineffective_at IS NULL; `
	err = r.DB.Select(&si, sq)
	if err != nil {
		// TODO: Change messaging if err is sql.NoRows then say couldn't find instance with instanceId
		log.Error(fmt.Sprintf("cfsbapi.Instances() ! %s", err))
	}
	r.DB.Close()
	return
}

func Remove() (err error) {
}

func (i *Instance) Lock() (err error) {
	// Acquire consul schedulerLock to aquire right to schedule tasks.
	key := fmt.Sprintf("rdpg/cfsb/instance/id/%s", i.Id)
	client, _ := consulapi.NewClient(consulapi.DefaultConfig())
	schedulerLock, err := client.LockKey(key)
	if err != nil {
		log.Error(fmt.Sprintf("scheduler.Schedule() Error Locking Scheduler Key %s ! %s", key, err))
		return
	}

	lockCh, err := schedulerLock.Lock(nil) // Acquire Consul K/V Lock
	if err != nil {
		log.Error(fmt.Sprintf("scheduler.Lock() Error Aquiring Scheduler Key lock %s ! %s", key, err))
		return
	}

	if lockCh == nil {
		err = fmt.Errorf(`Scheduler Lock not aquired.`)
	}

	return
}

func (i *Instance) Unlock() (err error) {
	if schedulerLock != nil {
		err = schedulerLock.Unlock()
	}
	return
}
