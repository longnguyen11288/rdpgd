package pg

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	pguri string
	DB    *sqlx.DB
)

func init() {
	pguri = os.Getenv("RDPG_PG_URI")
	if pguri == "" || pguri[0:13] != "postgresql://" {
		fmt.Printf("ERROR: RDPGAPI_PG_URI is not set correctly in the environment.\n")
		proc, _ := os.FindProcess(os.Getpid())
		proc.Signal(syscall.SIGTERM)
	}
}

func Open() (err error) {
	if DB == nil {
		DB, err = sqlx.Connect("postgres", pguri)
		if err != nil {
			fmt.Printf("ERROR: %s:\n%s\n", pguri, err)
			return err
		}
	}
	err = DB.Ping()
	if err != nil {
		DB.Close()
		return errors.New(fmt.Sprintf("ERROR: %s:\n%s\n", pguri, err))
	}
	return nil
}

func Close() {
	DB.Close()
}
