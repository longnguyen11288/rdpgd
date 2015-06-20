package cfsbapi

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/wayneeseguin/rdpg-agent/log"
	"github.com/wayneeseguin/rdpg-agent/rdpg"
)

type Instance struct {
	Id             string `db:"id"`
	InstanceId     string `db:"instance_id" json:"instance_id"`
	ServiceId      string `db:"service_id" json:"service_id"`
	PlanId         string `db:"plan_id" json:"plan_id"`
	OrganizationId string `db:"organization_id" json:"organization_id"`
	SpaceId        string `db:"space_id" json:"space_id"`
	Database       string `db:"dbname" json:"dbname"`
	User           string `db:"uname" json:"uname"`
	Pass           string `db:"pass" json:"pass"`
}

func NewInstance(instanceId, serviceId, planId, organizationId, spaceId string) (i *Instance, err error) {
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
	r := rdpg.New()
	sq := ` SELECT id, instance_id, service_id, plan_id, organization_id, space_id, dbname, uname, pass 
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
	r := rdpg.New()
	in := Instance{}
	sq := `SELECT id, instance_id, service_id, plan_id, organization_id, space_id, dbname, uname, pass FROM cfsbapi.instances WHERE instance_id=lower($1) LIMIT 1;`
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
	i.Pass = strings.ToLower(strings.Replace(rdpg.NewUUID().String(), "-", "", -1))
	r := rdpg.New()

	// TODO: Alter this logic based on "plan"
	err = r.CreateUser(i.User, i.Pass)
	if err != nil {
		log.Error(fmt.Sprintf("Instance#Provision(%s) CreateUser(%s) ! %s", i.InstanceId, i.User, err))
		return err
	}

	err = r.CreateDatabase(i.Database, i.User)
	if err != nil {
		log.Error(fmt.Sprintf("Instance#Provision(%s) CreateDatabase(%s,%s) ! %s", i.InstanceId, i.Database, i.User, err))
		return err
	}

	err = r.CreateReplicationGroup(i.Database)
	if err != nil {
		log.Error(fmt.Sprintf("Instance#Provision(%s) CreateReplicationGroup(%s) ! %s", i.InstanceId, i.Database, err))
		return err
	}

	r.OpenDB("rdpg")
	sq := `INSERT INTO cfsbapi.instances 
(instance_id, service_id, plan_id, organization_id, space_id, dbname, uname, pass)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8);
`
	_, err = r.DB.Query(sq, i.InstanceId, i.ServiceId, i.PlanId, i.OrganizationId, i.SpaceId, i.Database, i.User, i.Pass)
	if err != nil {
		log.Error(fmt.Sprintf(`Instance#Provision(%s) ! %s`, i.InstanceId, err))
	}

	hosts := r.Hosts()
	for _, host := range hosts {
		err := host.AdminAPI("PUT", "services/pgbouncer/configure")
		if err != nil {
			log.Error(fmt.Sprintf(`Instance#Provision(%s) %s ! %s`, i.InstanceId, host.Host, err))
		}
	}
	r.DB.Close()
	return nil
}

func (i *Instance) Remove() (err error) {
	r := rdpg.New()
	r.OpenDB("rdpg")
	_, err = r.DB.Exec(`UPDATE cfsbapi.instances SET ineffective_at = CURRENT_TIMESTAMP WHERE id=$1`, i.Id)
	if err != nil {
		log.Error(fmt.Sprintf("Instance#Remove(%s) ! %s", i.InstanceId, err))
	}

	time.Sleep(1) // Wait for the update to propigate to the other hosts.

	for _, host := range r.Hosts() {
		err := host.AdminAPI("PUT", "services/pgbouncer/configure")
		if err != nil {
			log.Error(fmt.Sprintf(`Instance#Provision(%s) %s ! %s`, i.InstanceId, host.Host, err))
		}
	}
	r.DB.Close()
	return
}

func (i *Instance) ExternalDNS() (dns string) {
	// TODO: Figure out where we'll store and retrieve the external DNS information
	r := rdpg.New()
	hosts := r.Hosts()
	// TODO: Import the external DNS host via env variable configuration.
	return hosts[0].Host + ":5432"
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
	r := rdpg.New()
	r.OpenDB("rdpg")
	si = []Instance{}
	// TODO: Move this into a versioned SQL Function.
	sq := `SELECT instance_id, service_id, plan_id, organization_id, space_id, dbname, uname, 'md5'||md5(cfsbapi.instances.pass||uname) as pass FROM cfsbapi.instances WHERE ineffective_at IS NULL; `
	err = r.DB.Select(&si, sq)
	if err != nil {
		// TODO: Change messaging if err is sql.NoRows then say couldn't find instance with instanceId
		log.Error(fmt.Sprintf("cfsbapi.Instances() ! %s", err))
	}
	r.DB.Close()
	return
}
