package cfsb

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func EmptyHandler(w http.ResponseWriter, r *http.Request) {

}

func TestAPIRequirements(t *testing.T) {
	// API Version Header is set test
	// basic_auth test, need username and password (Authentication :header) to do broker registrations
	// return 401 Unauthorized if credentials are not valid  test, auth only tested here
	// test when reject a request, response a 412 Precondition Failed message
	httpHandlerFunc := http.HandlerFunc(httpAuth(EmptyHandler))
	if r, err := http.NewRequest("GET", "", nil); err != nil {
		t.Errorf("%v", err)
	} else {
		recorder := httptest.NewRecorder()
		httpHandlerFunc.ServeHTTP(recorder, r)
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("returned %v. Expected %v.", recorder.Code, http.StatusUnauthorized)
		}
	}

	httpHandlerFunc = http.HandlerFunc(httpAuth(CatalogHandler))
	if r, err := http.NewRequest("GET", "http://cf:cf@127.0.0.1:8080/v2/catalog", nil); err != nil {
		t.Errorf("%v", err)
	} else {
		recorder := httptest.NewRecorder()
		httpHandlerFunc.ServeHTTP(recorder, r)
		if recorder.Code != http.StatusOK {
			t.Errorf("returned %v. Expected %v.", recorder.Code, http.StatusOK)
		}
	}

}
func TestFetchCatalog(t *testing.T) {
	// PUT,POST,DELETE all respond with http.StatusMethodNotAllowed
	// GET
}

func TestBinding(t *testing.T) {
	// PUT
	// DELETE
}

func TestInstance(t *testing.T) {
	// PUT
	// DELETE
}

func TestHealth(t *testing.T) {
	// GET /health/hapbpg
}
