package rdpg

import (
	"fmt"

	"database/sql"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/wayneeseguin/rdpg-agent/log"
)

func initSchema(db *sqlx.DB) (err error) {
	log.Debug(fmt.Sprintf("rdpg.initializeSchema() for %s", rdpgURI))
	// For each node connect to pgbdr and:
	//   CreatDatabase('rdpg','postgres')
	//   "ALTER USER postgres SUPERUSER CREATEDB CREATEROLE INHERIT"
	//   CreateReplicationGroup('rdpg')

	log.Debug(fmt.Sprintf("rdpg.InitializeSchema() %s", SQL["rdpg_extensions"]))
	if _, err = db.Exec(SQL["rdpg_extensions"]); err != nil {
		log.Error(fmt.Sprintf("rdpg.InitializeSchema(rdpg_extensions) %s\n", err))
		return err
	}

	log.Debug(fmt.Sprintf("rdpg.InitializeSchema() %s", SQL["rdpg_schemas"]))
	if _, err = db.Exec(SQL["rdpg_schemas"]); err != nil {
		log.Error(fmt.Sprintf("rdpg.InitializeSchema(create_table_services) %s\n", err))
	}

	// TODO: Check if table exists first and only run if it doesn't.
	log.Debug(fmt.Sprintf("rdpg.InitializeSchema() %s", SQL["create_table_cfsb_services"]))
	if _, err = db.Exec(SQL["create_table_cfsb_services"]); err != nil {
		log.Error(fmt.Sprintf("rdpg.InitializeSchema(create_table_cfsb_services) %s\n", err))
	}

	// TODO: Check if table exists first and only run if it doesn't.
	log.Debug(fmt.Sprintf("rdpg.InitializeSchema() %s", SQL["create_table_cfsb_plans"]))
	if _, err = db.Exec(SQL["create_table_cfsb_plans"]); err != nil {
		log.Error(fmt.Sprintf("rdpg.InitializeSchema(create_table_plans) %s\n", err))
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
