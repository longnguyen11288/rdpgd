package cfsbapi

import (
	"fmt"

	"github.com/wayneeseguin/rdpgd/log"
	"github.com/wayneeseguin/rdpgd/rdpg"
)

type Catalog struct {
	Services []Service `json:"services"`
}

func (c *Catalog) Fetch() (err error) {
	r := rdpg.NewRDPG()
	err = r.OpenDB("rdpg")
	if err != nil {
		log.Error(fmt.Sprintf("cfsbapi.Catalog#FetchFailed() fetching catalog from database: %s", err))
		return
	}
	defer r.DB.Close()

	err = r.DB.Select(&c.Services, `SELECT service_id,name,description,bindable FROM cfsbapi.services;`)
	if err != nil {
		log.Error(fmt.Sprintf("Catalog#Fetch() selecting from cfsbapi.services %s", err.Error()))
		return
	}

	// TODO: Account for plans being associated with a service.
	for i, _ := range c.Services {
		service := &c.Services[i]
		err = r.DB.Select(&service.Plans, `SELECT plan_id,name,description FROM cfsbapi.plans;`)
		if err != nil {
			log.Error(fmt.Sprintf("Catalog#Fetch() Service Plans %s", err.Error()))
			return
		}
		c.Services[i].Tags = []string{"rdpg", "postgresql"}
		// c.Services[i].Dashboard = DashboardClient{}
	}
	return
}
