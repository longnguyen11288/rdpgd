package adminapi

import (
	"fmt"
	"net/http"

	"github.com/wayneeseguin/rdpg-agent/log"
	"github.com/wayneeseguin/rdpg-agent/rdpg"
)

func Check(check string) (status int) {
	r := rdpg.New()
	err := r.OpenDB("rdpg")
	if err != nil {
		log.Error(fmt.Sprintf("Error opening ! %s", r.URI))
		return http.StatusInternalServerError
	}

	switch check {
	case "ha_pb_pg":
		var numHosts int
		r.DB.Get(&numHosts, "SELECT count(node_name) FROM bdr.bdr_nodes;")
		if numHosts < 3 {
			return http.StatusInternalServerError
		}
	default:
		return http.StatusInternalServerError
	}
	return http.StatusOK
}
