package health

import (
	//"fmt"
	"net/http"
	//"github.com/wayneeseguin/rdpg-agent/bdr"
	"github.com/wayneeseguin/rdpg-agent/pg"
)

func Check(check string) (status int) {
	switch check {
	case "hapbpg":
		if !hapbpgCheck() {
			return http.StatusInternalServerError
		}
	default:
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func hapbpgCheck() bool {
	var numNodes int

	pg.DB.Get(&numNodes, "SELECT count(node_name) FROM bdr.bdr_nodes;")

	/*
		bdrNode := bdr.Node{}
		bdrNodes := []bdr.Node{}
		rows, err := pg.DB.Queryx("SELECT node_name,node_local_dsn,node_init_from_dsn FROM bdr.bdr_nodes;")
		if err != nil {
			return false
		}
		if rows.
		for rows.Next() {
			err := rows.StructScan(&bdrNode)
			if err != nil {
				fmt.Printf("%s\n",err)
			}  else {
				bdrNodes = append(bdrNodes,bdrNode)
			}
		}
	*/
	if numNodes > 2 {
		return true
	} else {
		return false
	}
}
