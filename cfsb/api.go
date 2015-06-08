package cfsb

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/wayneeseguin/rdpg-agent/log"
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
	router.HandleFunc("/v2/service_instances/{instance_id}", httpAuth(InstanceHandler))
	CFSBMux.Handle("/", router)
	router.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{binding_id}", httpAuth(BindingHandler))

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
			log.Error(fmt.Sprintf("httpAuth(): Unhandled Authorization Type, Expected Basic"))
			http.Error(w, "Unhandled Authroization Type, Expected Basic\n", http.StatusBadRequest)
			return
		}
		payload, err := base64.StdEncoding.DecodeString(auth[1])
		if err != nil {
			log.Error(fmt.Sprintf("httpAuth(): Authorization Failed"))
			http.Error(w, "Authorization Failed\n", http.StatusUnauthorized)
			return
		}
		nv := strings.SplitN(string(payload), ":", 2)
		if (len(nv) != 2) || !isAuthorized(nv[0], nv[1]) {
			log.Error(fmt.Sprintf("httpAuth(): Authorization Failed"))
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
	log.Trace(fmt.Sprintf("%s /v2/catalog", request.Method))
	switch request.Method {
	case "GET":
		c := Catalog{}
		err := c.Fetch()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		jsonCatalog, err := json.Marshal(c)
		if err != nil {
			log.Error(fmt.Sprintf("%s /v2/catalog %s", request.Method, err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		} else {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusOK)
			w.Write(jsonCatalog)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, `{"status": %d,"description": "Allowed Methods: GET"}`, http.StatusMethodNotAllowed)
		return
	}
}

/*
(PI) PUT /v2/service_instances/:id
(RI) DELETE /v2/service_instances/:id
*/
func InstanceHandler(w http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	log.Trace(fmt.Sprintf("%s /v2/service_instances/:instance_id :: %+v", request.Method, vars))
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	switch request.Method {
	case "PUT":
		type instanceRequest struct {
			ServiceId      string `json:"service_id"`
			Plan           string `json:"plan_id"`
			OrganizationId string `json:"organization_guid"`
			SpaceId        string `json:"space_guid"`
		}
		ir := instanceRequest{}
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			log.Error(fmt.Sprintf("%s /v2/service_instances/:instance_id %s", request.Method, err))
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"status": %d,"description": %s}`, http.StatusInternalServerError, err)
			return
		}
		err = json.Unmarshal(body, &ir)
		if err != nil {
			log.Error(fmt.Sprintf("%s /v2/service_instances/:instance_id %s", request.Method, err))
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"status": %d,"description": %s}`, http.StatusInternalServerError, err)
			return
		}

		instance, err := NewInstance(
			vars["instance_id"],
			ir.ServiceId,
			ir.Plan,
			ir.OrganizationId,
			ir.SpaceId,
		)
		if err != nil {
			log.Error(fmt.Sprintf("%s /v2/service_instances/:instance_id %s", request.Method, err))
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"status": %d,"description": %s}`, http.StatusInternalServerError, err)
			return
		}

		err = instance.Provision()
		if err != nil {
			log.Error(fmt.Sprintf("%s /v2/service_instances/:instance_id %s", request.Method, err))
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"status": %d,"description": %s}`, http.StatusInternalServerError, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": %d,"description": "Instance Provisioned Successfully"}`, http.StatusOK)
		return
	case "DELETE":
		instance, err := FindInstance(vars["instance_id"])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"status": %d,"description": %s}`, http.StatusInternalServerError, err)
			return
		}
		err = instance.Remove()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"status": %d,"description": %s}`, http.StatusInternalServerError, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": %d,"description": "Successfully Deprovisioned %s"}`, http.StatusOK, instance.InstanceId)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, `{"status": %d,"description": "Allowed Methods: PUT, DELETE"}`, http.StatusMethodNotAllowed)
		return
	}
}

/*
(CB) PUT /v2/service_instances/:instance_id/service_bindings/:binding_id
(RB) DELETE /v2/service_instances/:instance_id/service_bindings/:binding_id
*/
func BindingHandler(w http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"status": %d,"description": %s}`, http.StatusInternalServerError, err)
		return
	}
	log.Trace(fmt.Sprintf("%s /v2/service_instances/:instance_id/service_bindings/:binding_id :: %+v :: %s", request.Method, vars, body))
	switch request.Method {
	case "PUT":
		binding, err := CreateBinding(vars["instance_id"], vars["binding_id"])
		if err != nil {
			log.Error(fmt.Sprintf("%s /v2/service_instances/:instance_id/service_bindings/:binding_id %s", request.Method, err))
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"status": %d,"description": %s}`, http.StatusInternalServerError, err)
			return
		}
		j, err := json.Marshal(binding)
		if err != nil {
			log.Error(fmt.Sprintf("%s /v2/service_instances/:instance_id/service_bindings/:binding_id %s", request.Method, err))
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"status": %d,"description": %s}`, http.StatusInternalServerError, err)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(j)
			return
		}
	case "DELETE":
		//err := RemoveBinding(vars["instance_id"], vars["binding_id"])
		//if err != nil {
		//  w.WriteHeader(http.StatusInternalServerError)
		//fmt.Fprintf(w, `{"status": %d,"description": %s}`, http.StatusInternalServerError, err)
		//}
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"status": %d,"description": "NOT YET IMPLEMENTED"}`, http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, `{"status": %d,"description": "Allowed Methods: PUT, DELETE"}`, http.StatusMethodNotAllowed)
		return
	}
}
