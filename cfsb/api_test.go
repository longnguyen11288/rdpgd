package cfsb

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIRequirements(t *testing.T) {
	// API Version Header is set test
	// basic_auth test, need username and password (Authentication :header) to do broker registrations
	// return 401 Unauthorized if credentials are not valid  test, auth only tested here
	// test when reject a request, response a 412 Precondition Failed message

	var getBasicAuthTests = []struct {
		username, password string
		status             int
	}{
		{"cf", "cf", 200},
		{"Aladdin", "open:sesame", 401},
		{"", "", 401},
		{"cf", "bala", 401},
		{"", "cf", 401},
	}

	httpHandlerFunc := http.HandlerFunc(httpAuth(CatalogHandler))
	for _, autt := range getBasicAuthTests {
		r, err := http.NewRequest("GET", "http://cf:cf@127.0.0.1:8080/v2/catalog", nil)
		r.SetBasicAuth(autt.username, autt.password)
		if err != nil {
			t.Errorf("%v", err)
		} else {
			recorder := httptest.NewRecorder()
			httpHandlerFunc.ServeHTTP(recorder, r)
			if recorder.Code != autt.status {
				t.Errorf("returned %v. Expected %v.", recorder.Code, autt.status)
			}
		}
	}

}
func TestFetchCatalog(t *testing.T) {
	// PUT,POST,DELETE all respond with http.StatusMethodNotAllowed
	// GET
	httpHandlerFunc := http.HandlerFunc(httpAuth(CatalogHandler))
	r, err := http.NewRequest("GET", "http://cf:cf@127.0.0.0.1:8080/v2/catalog", nil)
	r.SetBasicAuth("cf", "cf")
	if err != nil {
		t.Errorf("%v", err)
	} else {
		recorder := httptest.NewRecorder()
		httpHandlerFunc.ServeHTTP(recorder, r)
		if recorder.Code != http.StatusOK {
			t.Errorf("returned %v, Expected %v.", recorder.Code, http.StatusOK)
		}
	}
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
