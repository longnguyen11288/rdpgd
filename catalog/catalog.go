package catalog

import (
	"github.com/wayneeseguin/rdpg-agent/pg"
)

type Catalog struct {
}

func Fetch() (Catalog, error) {
	pg.DB.QueryX(&numNodes, `SELECT id,name,description,bindable,dashboard_client FROM pgbdr.services;`)

	pg.DB.QueryX(&numNodes, `SELECT id,service_id,name,description,free FROM pgbdr.plans WHERE service_id='';`)

	rows, err := db.Queryx("SELECT node_name,node_local_dsn,node_init_from_dsn FROM bdr.bdr_nodes;")
	for rows.Next() {
		err := rows.StructScan(&bdrNode)
		if err != nil {
			fmt.Printf("%s\n",err)
		}  else {
			fmt.Printf("%+v\n",bdrNode)
			bdrNodes = append(bdrNodes,bdrNode)
		}
	}

	
	return catalog, nil
}
