package cfsb

import (
	"fmt"

	"crypto/rand"

	"github.com/wayneeseguin/rdpg-agent/log"
	"github.com/wayneeseguin/rdpg-agent/rdpg"
)

type Instance struct {
	Id             string
	ServiceId      string
	PlanId         string
	OrganizationId string
	Database       string
	User           string
	Pass           string
}

func NewInstance(instanceId, serviceId, planId, organizationId, spaceId string) *Instance {
	return &Instance{
		Id:             instanceId,
		ServiceId:      serviceId,
		PlanId:         planId,
		OrganizationId: organizationId,
		Database:       instanceId,
		User:           instanceId,
	}
}

func (i *Instance) Provision() (err error) {
	b := make([]byte, 16)
	rand.Read(b)
	i.Pass = fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	r := rdpg.New()

	err = r.CreateUser(i.User, i.Pass)
	if err != nil {
		log.Error(fmt.Sprintf("Instance#Provision() %s", err))
		return err
	}

	err = r.CreateDatabase(i.Database, i.User)
	if err != nil {
		log.Error(fmt.Sprintf("Instance#Provision() %s", err))
		return err
	}

	err = r.CreateReplicationGroup(i.Database)
	if err != nil {
		log.Error(fmt.Sprintf("Instance#Provision() %s", err))
		return err
	}
	return nil
}

func (i *Instance) Remove() error {
	r := rdpg.New()
	r.DisableDatabase()
	r.BackupDatabase()
	r.DeleteDatabase()
	r.DeleteUser()
	return nil
}
