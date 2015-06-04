package cfsb

import (
	"fmt"

	"github.com/wayneeseguin/rdpg-agent/log"
	"github.com/wayneeseguin/rdpg-agent/rdpg"
)

type Catalog struct {
	Services []Service `json:"services"`
}

func (c *Catalog) Fetch() error {
	// select and populate the c.Services.
	r := rdpg.New()
	err := r.Open()
	if err != nil {
		log.Error(fmt.Sprintf("Failed fetching catalog from database: %s", err))
	}

	services := []Service{}
	err = r.DB.Select(&services, `SELECT id,name,description,bindable,dashboard_client FROM cfsb.services;`)
	if err != nil {
		log.Error(fmt.Sprintf("Catalog#Fetch() selecting from cfsb.services %s", err.Error()))
	}

	//pg.DB.QueryX(&numNodes, `SELECT id,service_id,name,description,free FROM pgbdr.plans WHERE service_id='';`)

	/*
		n := pgbdr.NewNode("127.0.0.1","5432","postgres","rdpg")
		for rows.Next() {
			err := rows.StructScan(&c)
			if err != nil {
				fmt.Printf("%s\n",err)
			}  else {
				fmt.Printf("%+v\n",n)
			}
		}
	*/
	return nil
}
