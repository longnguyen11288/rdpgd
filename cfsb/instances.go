package cfsb

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/wayneeseguin/rdpg-agent/log"
	"github.com/wayneeseguin/rdpg-agent/rdpg"
)

type Instance struct {
	Id             int    `db:"id"`
	InstanceId     string `db:"instance_id"`
	ServiceId      string `db:"service_id"`
	PlanId         string `db:"plan_id"`
	OrganizationId string `db:"organization_id"`
	SpaceId        string `db:"space_id"`
	Database       string `db:"dbname"`
	User           string `db:"uname"`
	Pass           string `db:"pass"`
}

func NewInstance(instanceId, serviceId, planId, organizationId, spaceId string) *Instance {
	re := regexp.MustCompile("[^A-Za-z0-9_]")
	identifier := strings.ToLower(string(re.ReplaceAll([]byte(instanceId), []byte(""))))

	return &Instance{
		InstanceId:     strings.ToLower(instanceId),
		ServiceId:      strings.ToLower(serviceId),
		PlanId:         strings.ToLower(planId),
		OrganizationId: strings.ToLower(organizationId),
		Database:       "d" + identifier,
		User:           "u" + identifier,
	}
}

func FindInstance(instanceId string) (i *Instance, err error) {
	r := rdpg.New()
	r.Open()
	i = &Instance{}
	sq := `SELECT 
id, instance_id, service_id, plan_id, organization_id, space_id, dbname, uname, pass 
FROM cfsb.instances WHERE instance_id=$1 LIMIT 1;`
	err = r.DB.Get(&i, sq, instanceId)
	if err != nil {
		log.Error(fmt.Sprintf("cfsb.FindInstance(%s) %s\n", instanceId, err))
	}
	r.DB.Close()
	return i, err
}

func (i *Instance) Provision() (err error) {
	i.Pass = strings.ToLower(strings.Replace(rdpg.NewUUID().String(), "-", "", -1))
	r := rdpg.New()

	err = r.CreateUser(i.User, i.Pass)
	if err != nil {
		log.Error(fmt.Sprintf("Instance#Provision() %s\n", err))
		return err
	}

	err = r.CreateDatabase(i.Database, i.User)
	if err != nil {
		log.Error(fmt.Sprintf("Instance#Provision() %s\n", err))
		return err
	}

	err = r.CreateReplicationGroup(i.Database)
	if err != nil {
		log.Error(fmt.Sprintf("Instance#Provision() %s\n", err))
		return err
	}

	r.Open()
	sq := `INSERT INTO cfsb.instances 
(instance_id, service_id, plan_id, organization_id, space_id, dbname, uname, pass)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8);
`
	_, err = r.DB.Query(sq, i.Id, i.ServiceId, i.PlanId, i.OrganizationId, i.SpaceId, i.Database, i.User, i.Pass)
	if err != nil {
		log.Error(fmt.Sprintf(`Instance#Provision() %s\n`, err))
	}

	return nil
}

func (i *Instance) Remove() error {
	r := rdpg.New()
	r.DisableDatabase(i.Database)
	r.BackupDatabase(i.Database)
	r.DropDatabase(i.Database)
	r.DropUser(i.User)

	// TODO: Once all database have been nuked, delete the instance.
	// db.Exec("UPDATE cfsb.instances SET ineffective_at = CURRENT_TIMESTAMP WHERE id=$1", dbname);)

	return nil
}

func (i *Instance) ExternalDNS() (dns string) {
	// TODO: Figure out where we'll store and retrieve the external DNS information
	r := rdpg.New()
	nodes := r.Nodes()
	return nodes[0].Host + ":" + nodes[0].Port
}

func (i *Instance) URI() (uri string) {
	d := `postgres://%s@%s/%s?connect_timeout=%s&sslmode=%s`
	// TODO:
	uri = fmt.Sprintf(d, i.User, i.ExternalDNS(), i.Database, `5`, `disable`)
	return
}

func (i *Instance) DSN() (uri string) {
	dns := i.ExternalDNS()
	s := strings.Split(dns, ":")
	d := `user=%s host=%s port=%s dbname=%s connect_timeout=%s sslmode=%s`
	uri = fmt.Sprintf(d, i.User, s[0], s[1], i.Database, `5`, `disable`)
	return
}

func (i *Instance) JDBCURI() (uri string) {
	dns := i.ExternalDNS()
	s := strings.Split(dns, ":")
	d := `user=%s host=%s port=%s dbname=%s connect_timeout=%s sslmode=%s`
	uri = fmt.Sprintf(d, i.User, s[0], s[1], i.Database, `5`, `disable`)
	return
}
