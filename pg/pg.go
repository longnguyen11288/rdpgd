package pg

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
)

var (
	pguri string
	DB    *sqlx.DB
)

func init() {
	pguri = os.Getenv("RDPGAPI_PGURI")
	if pguri == "" || pguri[0:13] != "postgresql://" {
		fmt.Printf("ERROR: RDPGAPI_PGURI is not set correctly in the environment.\n")
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
		return errors.New(fmt.Sprintf("ERROR: %s:\n%s\n", pguri,err))
	}
	return nil
}

func Close() {
	DB.Close()
}
