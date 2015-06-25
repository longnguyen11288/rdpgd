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
		var numNodes int
		r.DB.Get(&numNodes, "SELECT count(node_name) FROM bdr.bdr_nodes;")
		if numNodes < 2 {
			return http.StatusInternalServerError
		}
	case "pg":
		_, err = r.DB.Exec(`SELECT CURRENT_TIMESTAMP;`)
		if err != nil {
			return http.StatusInternalServerError
		}
	default:
		return http.StatusInternalServerError
	}
	return http.StatusOK
}
