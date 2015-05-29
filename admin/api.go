package admin

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/wayneeseguin/rdpg-agent/health"
)

var (
	port, sbUser, sbPass string
)

// StatusPreconditionFailed
func init() {
	port = os.Getenv("RDPGAPI_ADMIN_PORT")
	if port == "" {
		port = "8888"
	}
	sbUser = os.Getenv("RDPGAPI_ADMIN_USER")
	if sbUser == "" {
		sbUser = "admin"
	}
	sbPass = os.Getenv("RDPGAPI_ADMIN_PASS")
	if sbPass == "" {
		sbPass = "admin"
	}
}

func API() {
	router := mux.NewRouter()

	router.HandleFunc("/health/{check}", auth(Health))

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
