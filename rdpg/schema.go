package rdpg

import (
	"fmt"

	"database/sql"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/wayneeseguin/rdpg-agent/log"
)

func initSchema(db *sqlx.DB) (err error) {
	log.Trace(fmt.Sprintf("rdpg.initializeSchema() for %s", rdpgURI))
	// TODO: if 'rdpg' database DNE,
	// For each node connect to pgbdr and:
	//   CreatDatabase('rdpg','postgres')
	//   "ALTER USER postgres SUPERUSER CREATEDB CREATEROLE INHERIT"
	//   CreateReplicationGroup('rdpg')

	keys := []string{
		"rdpg_extensions",
		"rdpg_schemas",
		"create_table_cfsb_services",
		"create_table_cfsb_plans",
		"create_table_cfsb_instances",
		"create_table_cfsb_bindings",
		"create_table_cfsb_credentials",
	}
	// "create_table_rdpg_nodes"
	// "create_table_monitoring_...
	// TODO: Check if table exists first and only run if it doesn't.
	for _, key := range keys {
		log.Trace(fmt.Sprintf("RDPG#InitializeSchema() SQL[%s]", key))
		_, err = db.Exec(SQL[key])
		if err != nil {
			log.Error(fmt.Sprintf("RDPG#InitializeSchema() %s\n", err))
		}
	}

	var name string

	// TODO: Move initial population of services out of rdpg-agent to Admin API.
	if err := db.QueryRow("SELECT name FROM cfsb.services WHERE name='rdpg' LIMIT 1;").Scan(&name); err != nil {
		if err == sql.ErrNoRows {
			if _, err = db.Exec(SQL["insert_default_cfsb_services"]); err != nil {
				log.Error(fmt.Sprintf("rdpg.InitializeSchema(insert_default_cfsb_services) %s\n", err))
				return err
			}
		} else {
			log.Error(fmt.Sprintf("rdpg.InitializeSchema() %s\n", err))
			return err
		}
	}

	// TODO: Move initial population of services out of rdpg-agent to Admin API.
	if err = db.QueryRow("SELECT name FROM cfsb.plans WHERE name='small' LIMIT 1;").Scan(&name); err != nil {
		if err == sql.ErrNoRows {
			if _, err = db.Exec(SQL["insert_default_cfsb_plans"]); err != nil {
				log.Error(fmt.Sprintf("rdpg.InitializeSchema(insert_default_cfsb_plans) %s\n", err))
				return err
			}
		} else {
			log.Error(fmt.Sprintf("rdpg.InitializeSchema() %s\n", err))
			return err
		}
	}

	return nil
}
