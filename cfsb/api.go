package cfsb

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

var (
	port, sbUser, sbPass string
)

// StatusPreconditionFailed
func init() {
	port = os.Getenv("RDPGAPI_SB_PORT")
	if port == "" {
		port = "8080"
	}
	sbUser = os.Getenv("RDPGAPI_SB_USER")
	if sbUser == "" {
		sbUser = "cf"
	}
	sbPass = os.Getenv("RDPGAPI_SB_PASS")
	if sbPass == "" {
		sbPass = "cf"
	}
}

func API() {
	router := mux.NewRouter()

	router.HandleFunc("/v2/catalog", auth(FetchCatalog))
	router.HandleFunc("/v2/service_instances/{id}", auth(Instance))
	router.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{id}", auth(Binding))

	http.Handle("/", router)
	http.ListenAndServe(":"+port, nil)
}

func auth(h http.HandlerFunc) http.HandlerFunc {
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
		if (len(nv) != 2) || ! isAuthorized(nv[0], nv[1]) {
			http.Error(w, "Authorization Failed\n", http.StatusUnauthorized)
			return
		}
		h(w, request)
	}
}

func isAuthorized(username, password string) (bool) {
	if username == sbUser && password == sbPass {
		return true
	}
	return false
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

