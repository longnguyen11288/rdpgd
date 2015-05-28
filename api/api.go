package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/wayneeseguin/rdpg-agent/health"
	"github.com/wayneeseguin/rdpg-agent/catalog"
	//"github.com/wayneeseguin/rdpg-agent/plans"
	//"github.com/wayneeseguin/rdpg-agent/services"
	//"github.com/wayneeseguin/rdpg-agent/instances"
	//"github.com/wayneeseguin/rdpg-agent/bindings"
)

var port string

func init() {
	port = os.Getenv("RDPGAPI_PORT")
	if port == "" {
		port = "8080"
	}
}

func Run() {
	router := mux.NewRouter()
	RegisterEndpoints(router)
	http.Handle("/", router)
	http.ListenAndServe(":"+port, nil)
}

func RegisterEndpoints(r *mux.Router) {
	r.HandleFunc("/v2/catalog", FetchCatalog)
	r.HandleFunc("/v2/service_instances/{id}", Instance)
	r.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{id}", Binding)
	r.HandleFunc("/health/{check}", Health)
}

/*
(FC) GET /v2/catalog
*/
func FetchCatalog(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("X-Broker-Api-Version", "2.4")
	switch request.Method {
	case "GET":
		cat, err := catalog.Catalog()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		jsonCatalog, err := json.Marshal(cat)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(jsonCatalog)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "{}")
	}
}

/*
(PI) PUT /v2/service_instances/:id
(RI) DELETE /v2/service_instances/:id
*/
func Instance(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("X-Broker-Api-Version", "2.4")
	switch request.Method {
	case "PUT": // Provision Instance
		fmt.Fprintf(w, "{}")
	case "DELETE": // Remove Instance
		fmt.Fprintf(w, "{}")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "{}")
	}
}

/*
(CB) PUT /v2/service_instances/:instance_id/service_bindings/:id
(RB) DELETE /v2/service_instances/:instance_id/service_bindings/:id
*/
func Binding(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("X-Broker-Api-Version", "2.4")
	switch request.Method {
	case "PUT":
		// binding.Create(instance_id,binding_id)
		fmt.Fprintf(w, "{}")
	case "DELETE":
		// binding.Remove(instance_id,binding_id)
		fmt.Fprintf(w, "{}")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "{}")
	}
}

/*
(HC) GET /health/hapbpg
*/
func Health(w http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		vars := mux.Vars(request)
		w.WriteHeader(health.Check(vars["check"]))
		// health check...
		fmt.Fprintf(w, "{}")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "{}")
	}
}

