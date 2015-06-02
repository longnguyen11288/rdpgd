package admin

import (
	"net/http"

	"github.com/wayneeseguin/rdpg-agent/pg"
)

func Check(check string) (status int) {
	switch check {
	case "ha_pb_pg":
		var numNodes int
		pg.DB.Get(&numNodes, "SELECT count(node_name) FROM bdr.bdr_nodes;")
		if numNodes < 3 {
			return http.StatusInternalServerError
		}
	default:
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

