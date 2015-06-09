package admin

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/wayneeseguin/rdpg-agent/log"
)

var (
	adminPort, adminUser, adminPass string
)

type Admin struct {
}

func init() {
	adminPort = os.Getenv("RDPG_ADMIN_PORT")
	if adminPort == "" {
		adminPort = "58888"
	}
	adminUser = os.Getenv("RDPG_ADMIN_USER")
	if adminUser == "" {
		adminUser = "admin"
	}
	adminPass = os.Getenv("RDPG_ADMIN_PASS")
	if adminPass == "" {
		adminPass = "admin"
	}
}

func API() {
	AdminMux := http.NewServeMux()
	router := mux.NewRouter()
	router.HandleFunc("/health/{check}", httpAuth(HealthHandler))
	router.HandleFunc(`/services/{service}/{action}`, httpAuth(ServiceHandler))
	AdminMux.Handle("/", router)
	http.ListenAndServe(":"+adminPort, AdminMux)
}

func httpAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		if len(request.Header["Authorization"]) == 0 {
			log.Trace(fmt.Sprintf("httpAuth(): Authorization Required"))
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

/*
POST /services/{service}/{action}
*/
func ServiceHandler(w http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	log.Trace(fmt.Sprintf("%s /services/%s/%s", request.Method, vars["service"], vars["action"]))
	switch request.Method {
	case "PUT":
		service, err := NewService(vars["service"])
		if err != nil {
			log.Error(fmt.Sprintf("ServiceHandler(): NewService(%s)"))
			http.Error(w, `{"status": %d, "description": "%s"}`, http.StatusInternalServerError)
			return
		}

		switch vars["action"] {
		case "configure":
			err := service.Configure()
			if err != nil {
				msg := fmt.Sprintf(`{"status": %d, "description": "%s"}`, http.StatusInternalServerError, err)
				log.Error(msg)
				http.Error(w, msg, http.StatusInternalServerError)
			}
			msg := fmt.Sprintf(`{"status": %d, "description": "%s %s"}`, http.StatusOK, vars["service"], vars["action"])
			log.Trace(msg)
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, msg)
		default:
			msg := fmt.Sprintf(`{"status": %d, "description": "Invalid Action %s for %s"}`, http.StatusBadRequest, vars["action"], vars["service"])
			log.Error(msg)
			http.Error(w, msg, http.StatusBadRequest)
		}
	default:
		msg := fmt.Sprintf(`{"status": %d, "description": "Method not allowed %s"}`, http.StatusMethodNotAllowed, request.Method)
		log.Error(msg)
		http.Error(w, msg, http.StatusMethodNotAllowed)
	}
}
