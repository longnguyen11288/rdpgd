package api

import (
	"net/http"
	"fmt"
	"os"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/wayneeseguin/rdpg-agent/catalog"
	//"github.com/wayneeseguin/rdpg-agent/sb"
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
	r.HandleFunc("/plans/register/{cf_host_port}", Plans)
}

func Run() {
	router := mux.NewRouter()
	RegisterEndpoints(router)
	http.Handle("/", router)
	http.ListenAndServe(":" + port, nil)
}

// Cloud Controller (final release v145+) authenticates with the Broker using HTTP basic authentication (the Authorization: header) on every request and will reject any broker registrations that do not contain a username and password. The broker is responsible for checking the username and password and returning a 401 Unauthorized message if credentials are invalid. Cloud Controller supports connecting to a broker using SSL if additional security is desired.
// When Cloud Controller fetches a catalog from a broker, it will compare the broker’s id for services and plans with the unique_id values for services and plan in the Cloud Controller database. If a service or plan in the broker catalog has an id that is not present amongst the unique_id values in the database, a new record will be added to the database. If services or plans in the database are found with unique_ids that match the broker catalog’s id, Cloud Controller will update update the records to match the broker’s catalog.
// If the database has plans which are not found in the broker catalog, and there are no associated service instances, Cloud Controller will remove these plans from the database. Cloud Controller will then delete services that do not have associated plans from the database. If the database has plans which are not found in the broker catalog, and there are provisioned instances, the plan will be marked “inactive” and will no longer be visible in the marketplace catalog or be provisionable.

//(FC)GET /v2/catalog
func FetchCatalog(w http.ResponseWriter, request *http.Request) { 
	w.Header().Set("X-Broker-Api-Version", "2.4")
	switch request.Method {
	default:
		fmt.Fprintf(w, "{}")
	case "GET":
		cat,err := catalog.Catalog()
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

//(PI)PUT /v2/service_instances/:id
//(RI)DELETE /v2/service_instances/:id
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

//(CB)PUT /v2/service_instances/:instance_id/service_bindings/:id
//(RB)DELETE /v2/service_instances/:instance_id/service_bindings/:id
func Binding(w http.ResponseWriter, request *http.Request) { 
	w.Header().Set("X-Broker-Api-Version", "2.4")
	switch request.Method {
	default:
		fmt.Fprintf(w, "{}")
	case "PUT": // CreateBinding
		fmt.Fprintf(w, "{}")
	case "DELETE": // RemoveBinding
		fmt.Fprintf(w, "{}")
	}
}

//(RP)PUT /plans/register/:cf_host_port
//(RP)DELETE /plans/register/:cf_host_port
func Plans(w http.ResponseWriter, request *http.Request) { 
	w.Header().Set("X-Broker-Api-Version", "2.4")
	switch request.Method {
	default:
		fmt.Fprintf(w, "{}")
	case "PUT":
		fmt.Fprintf(w, "{}")
		case "DELETE": // Deregister
		/*
		err := b.DeletePlan(params["service_id"], params["plan_id"])

		if err == nil {
			if registerCFServices {
				// We need to specifically set the service broker password again, due to a
				// possible bug in the CF client library.
				sb.Password = brokerConfig.CFPassword
				fmt.Println(cfs.UpdateCFService(sb))
			}
			return http.StatusOK, "{}"
		}
		return http.StatusInternalServerError, MarshalError(err)
	*/
		fmt.Fprintf(w, "{}")
	}
}

