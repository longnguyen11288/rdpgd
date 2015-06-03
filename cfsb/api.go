package cfsb

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/wayneeseguin/rdpg-agent/log"
)

var (
	sbPort, sbUser, sbPass string
)

type CFSB struct {
}

// StatusPreconditionFailed
func init() {
	sbPort = os.Getenv("RDPG_SB_PORT")
	if sbPort == "" {
		sbPort = "8080"
	}
	sbUser = os.Getenv("RDPG_SB_USER")
	if sbUser == "" {
		sbUser = "cf"
	}
	sbPass = os.Getenv("RDPG_SB_PASS")
	if sbPass == "" {
		sbPass = "cf"
	}
}

func API() {
	CFSBMux := http.NewServeMux()
	router := mux.NewRouter()
	router.HandleFunc("/v2/catalog", httpAuth(CatalogHandler))
	router.HandleFunc("/v2/service_instances/{id}", httpAuth(InstanceHandler))
	CFSBMux.Handle("/", router)
	router.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{id}", httpAuth(BindingHandler))

	http.Handle("/", router)
	http.ListenAndServe(":"+sbPort, CFSBMux)
}

func httpAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		if len(request.Header["Authorization"]) == 0 {
			http.Error(w, "Authorization Required", http.StatusUnauthorized)
			return
		}

		auth := strings.SplitN(request.Header["Authorization"][0], " ", 2)
		if len(auth) != 2 || auth[0] != "Basic" {
			http.Error(w, "Unhandled Authroization Type, Expected Basic\n", http.StatusBadRequest)
			return
		}
		payload, err := base64.StdEncoding.DecodeString(auth[1])
		if err != nil {
			http.Error(w, "Authorization Failed\n", http.StatusUnauthorized)
			return
		}
		nv := strings.SplitN(string(payload), ":", 2)
		if (len(nv) != 2) || !isAuthorized(nv[0], nv[1]) {
			http.Error(w, "Authorization Failed\n", http.StatusUnauthorized)
			return
		}
		h(w, request)
	}
}

func isAuthorized(username, password string) bool {
	if username == sbUser && password == sbPass {
		return true
	}
	return false
}

/*
(FC) GET /v2/catalog
*/
func CatalogHandler(w http.ResponseWriter, request *http.Request) {
	// r.Header().Get("X-Broker-Api-Version") #  "2.4")

	switch request.Method {
	case "GET":
		c := Catalog{}
		err := c.Fetch()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		jsonCatalog, err := json.Marshal(c)
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
func InstanceHandler(w http.ResponseWriter, request *http.Request) {
	// r.Header().Get("X-Broker-Api-Version") #  "2.4")
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
func BindingHandler(w http.ResponseWriter, request *http.Request) {
	// r.Header().Get("X-Broker-Api-Version") #  "2.4")
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
