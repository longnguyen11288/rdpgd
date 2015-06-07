package admin

import (
	"fmt"
	"net/http"

	"github.com/wayneeseguin/rdpg-agent/log"
	"github.com/wayneeseguin/rdpg-agent/rdpg"
)

func Check(check string) (status int) {
	r := rdpg.New()
	err := r.OpenDB()
	if err != nil {
		log.Error(fmt.Sprintf("Error opening %s", r.URI))
		return http.StatusInternalServerError
	}

	switch check {
	case "ha_pb_pg":
		var numNodes int
		r.DB.Get(&numNodes, "SELECT count(node_name) FROM bdr.bdr_nodes;")
		if numNodes < 3 {
			return http.StatusInternalServerError
		}
	default:
		return http.StatusInternalServerError
	}
	return http.StatusOK
}
