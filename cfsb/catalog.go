package cfsb

import (
	//"fmt"
	//"github.com/wayneeseguin/rdpg-agent/pg"
	//"github.com/wayneeseguin/rdpg-agent/pgbdr"
)

type Catalog struct {
}

func (c *Catalog) Fetch() (error) {
	//pg.DB.QueryX(&numNodes, `SELECT id,name,description,bindable,dashboard_client FROM pgbdr.services;`)

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
