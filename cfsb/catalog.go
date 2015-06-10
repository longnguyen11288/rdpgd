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

func (c *Catalog) Fetch() (err error) {
	r := rdpg.New()
	err = r.OpenDB("rdpg")
	if err != nil {
		log.Error(fmt.Sprintf("Failed fetching catalog from database: %s", err))
		return
	}
	db := r.DB

	err = db.Select(&c.Services, `SELECT service_id,name,description,bindable FROM cfsb.services;`)
	if err != nil {
		log.Error(fmt.Sprintf("Catalog#Fetch() selecting from cfsb.services %s", err.Error()))
		return
	}

	// TODO: Account for plans being associated with a service.
	for i, _ := range c.Services {
		service := &c.Services[i]
		err = db.Select(&service.Plans, `SELECT plan_id,name,description FROM cfsb.plans;`)
		if err != nil {
			log.Error(fmt.Sprintf("Catalog#Fetch() Service Plans %s", err.Error()))
			return
		}
		c.Services[i].Tags = []string{"rdpg", "postgresql"}
		// c.Services[i].Dashboard = DashboardClient{}
	}
	return
}
