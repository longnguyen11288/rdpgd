package cfsb

import (
	"fmt"

	//. "github.com/smartystreets/goconvey/convey"

	"github.com/wayneeseguin/rdpg-agent/log"
	"github.com/wayneeseguin/rdpg-agent/rdpg"
)

type Catalog struct {
	Services []Service `json:"services"`
}

func (c *Catalog) Fetch() error {
	r := rdpg.New()
	err := r.OpenDB()
	if err != nil {
		log.Error(fmt.Sprintf("Failed fetching catalog from database: %s", err))
		return err
	}

	err = r.DB.Select(&c.Services, `SELECT service_id,name,description,bindable,dashboard_client FROM cfsb.services;`)
	if err != nil {
		log.Error(fmt.Sprintf("Catalog#Fetch() selecting from cfsb.services %s", err.Error()))
		return err
	}

	for i, _ := range c.Services {
		service := &c.Services[i]
		err = r.DB.Select(&service.Plans, `SELECT plan_id,name,description FROM cfsb.plans;`)
		if err != nil {
			log.Error(fmt.Sprintf("Catalog#Fetch() selectiing from cfsb.plans %s", err.Error()))
			return err
		}
	}
	return nil
}
