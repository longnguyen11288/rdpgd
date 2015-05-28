package health

import(
	"net/http"
)

func Check(check string) (status int){
	switch check  {
	case "hapbpg":
		// TODO: connect through haproxy <> pgbouncer <> pgbdr
		// if select true {
		// return http.StatusOK
		// }
	}
	return http.StatusInternalServerError
	/* TODO: Connect through */
}
