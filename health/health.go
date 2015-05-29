package health

import (
	"net/http"
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

	if numNodes > 2 {
		return true
	} else {
		return false
	}
}
