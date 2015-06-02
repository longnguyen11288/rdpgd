package admin

import (
	"encoding/base64"
	//"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/wayneeseguin/rdpg-agent/health"
)

var (
	adminPort, adminUser, adminPass string
)

// StatusPreconditionFailed
func init() {
	adminPort = os.Getenv("RDPGAPI_ADMIN_PORT")
	if adminPort == "" {
		adminPort = "58888"
	}
	adminUser = os.Getenv("RDPGAPI_ADMIN_USER")
	if adminUser == "" {
		adminUser = "admin"
	}
	adminPass = os.Getenv("RDPGAPI_ADMIN_PASS")
	if adminPass == "" {
		adminPass = "admin"
	}
}

func API() {
	router := mux.NewRouter()

	router.HandleFunc("/health/{check}", httpAuth(HealthHandler))

	http.Handle("/", router)
	http.ListenAndServe(":"+adminPort, nil)
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
		if (len(nv) != 2) || ! isAuthorized(nv[0], nv[1]) {
			http.Error(w, "Authorization Failed\n", http.StatusUnauthorized)
			return
		}
		h(w, request)
	}
}

func isAuthorized(username, password string) (bool) {
	if username == adminUser && password == adminPass {
		return true
	}
	return false
}

/*
(HC) GET /health/hapbpg
*/
func HealthHandler(w http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "GET":
		vars := mux.Vars(request)
		w.WriteHeader(Check(vars["check"]))
		fmt.Fprintf(w, "{}")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "{}")
	}
}
