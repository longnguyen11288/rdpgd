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
	r := rdpg.New()
	err := r.Open()
	if err != nil {
		log.Error(fmt.Sprintf("Failed fetching catalog from database: %s", err))
		return err
	}

	err = r.DB.Select(&c.Services, `SELECT id,name,description,bindable,dashboard_client FROM cfsb.services;`)
	if err != nil {
		log.Error(fmt.Sprintf("Catalog#Fetch() selecting from cfsb.services %s", err.Error()))
		return err
	}
	return nil
}
