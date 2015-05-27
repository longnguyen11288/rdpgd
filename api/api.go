package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
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

func RegisterEndpoints(r *mux.Router) {
	r.HandleFunc("/v2/catalog", FetchCatalog)
	r.HandleFunc("/v2/service_instances/{id}", Instance)
	r.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{id}", Binding)
}

func Run() {
	router := mux.NewRouter()
	RegisterEndpoints(router)
	http.Handle("/", router)
	http.ListenAndServe(":"+port, nil)
}

/*
(FC)GET /v2/catalog
*/
func FetchCatalog(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("X-Broker-Api-Version", "2.4")
	switch request.Method {
	default:
		fmt.Fprintf(w, "{}")
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
	}
}

/*
(PI)PUT /v2/service_instances/:id
(RI)DELETE /v2/service_instances/:id
*/
func Instance(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("X-Broker-Api-Version", "2.4")
	//vars := mux.Vars(request)
	//id := vars["id"]
	switch request.Method {
	default:
		fmt.Fprintf(w, "{}")
	case "PUT": // Provision Instance
		/*
			requestParams := RequestParams{}
			body, _ := ioutil.ReadAll(request.Body)
			json.Unmarshal(body,&requestParams)

			_, err := b.Provision(requestParams.PlanId, id)
			if err == nil {
				return http.StatusCreated, "{}"
			}
			return http.StatusInternalServerError, MarshalError(err)

		*/
		fmt.Fprintf(w, "{}")
	case "DELETE": // Remove Instance
		//planId := request.URL.Query()["plan_id"][0]

		//err := b.Deprovision(planId, params["id"])

		// if err == nil { return http.StatusOK, "{}" ; }
		//return http.StatusInternalServerError, MarshalError(err)
		fmt.Fprintf(w, "{}")
	}
}

/*
(CB)PUT /v2/service_instances/:instance_id/service_bindings/:id
(RB)DELETE /v2/service_instances/:instance_id/service_bindings/:id
*/
func Binding(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("X-Broker-Api-Version", "2.4")
	switch request.Method {
	default:
		fmt.Fprintf(w, "{}")
	case "PUT":
		// binding.Create()
		fmt.Fprintf(w, "{}")
	case "DELETE":
		// binding.Remove()
		fmt.Fprintf(w, "{}")
	}
}
