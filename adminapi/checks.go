package adminapi

import (
	"fmt"
	"net/http"

	"github.com/wayneeseguin/rdpgd/log"
	"github.com/wayneeseguin/rdpgd/rdpg"
)

func Check(check string) (status int) {
	r := rdpg.NewRDPG()
	err := r.OpenDB("rdpg")
	if err != nil {
		log.Error(fmt.Sprintf("adminapi.Check() %s ! %s", r.URI, err))
		return http.StatusInternalServerError
	}
	defer r.DB.Close()

	switch check {
	case "ha_pb_pg":
		var numHosts int
		r.DB.Get(&numHosts, "SELECT count(node_name) FROM bdr.bdr_nodes;")
		if numHosts < 2 {
			return http.StatusInternalServerError
		}
	default:
		return http.StatusInternalServerError
	}
	return http.StatusOK
}
