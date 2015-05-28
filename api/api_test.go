package api
import (
	"testing"
)

func TestAPIRequirements(t *testing.T) {
	// API Version Header is set test
	// basic_auth test, need username and password (Authentication :header) to do broker registrations
	// return 401 Unauthorized if credentials are not valid  test, auth only tested here
	// test when reject a request, response a 412 Precondition Failed message
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
