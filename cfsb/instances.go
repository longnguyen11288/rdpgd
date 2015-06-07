package cfsb

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/wayneeseguin/rdpg-agent/log"
	"github.com/wayneeseguin/rdpg-agent/rdpg"
)

type Instance struct {
	Id             string
	ServiceId      string
	PlanId         string
	OrganizationId string
	SpaceId        string
	Database       string
	User           string
	Pass           string
}

func NewInstance(instanceId, serviceId, planId, organizationId, spaceId string) *Instance {
	re := regexp.MustCompile("[^A-Za-z0-9_]")
	identifier := strings.ToLower(string(re.ReplaceAll([]byte(instanceId), []byte(""))))

	return &Instance{
		Id:             strings.ToLower(instanceId),
		ServiceId:      strings.ToLower(serviceId),
		PlanId:         strings.ToLower(planId),
		OrganizationId: strings.ToLower(organizationId),
		Database:       "d" + identifier,
		User:           "u" + identifier,
	}
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
