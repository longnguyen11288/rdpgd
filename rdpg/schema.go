package rdpg

import (
	"fmt"
	"strings"

	"database/sql"

	_ "github.com/lib/pq"
	"github.com/wayneeseguin/rdpg-agent/log"
)

// TODO: This should only be run on one node...
func (r *RDPG) InitSchema() (err error) {
	log.Trace(fmt.Sprintf("RDPG#initSchema() for %s", rdpgURI))
	// TODO: if 'rdpg' database DNE,
	// For each node connect to pgbdr and:
	//   CreatDatabase('rdpg','postgres')
	//   "ALTER USER postgres SUPERUSER CREATEDB CREATEROLE INHERIT"
	//   CreateReplicationGroup('rdpg')

	var name string

	db := r.DB

	keys := []string{
		"rdpg_extensions",
		"rdpg_schemas",
	}
	for _, key := range keys {
		log.Trace(fmt.Sprintf("RDPG#initSchema() SQL[%s]", key))
		_, err = db.Exec(SQL[key])
		if err != nil {
			log.Error(fmt.Sprintf("RDPG#initSchema() %s", err))
		}
	}

	keys = []string{
		"create_table_cfsb_services",
		"create_table_cfsb_plans",
		"create_table_cfsb_instances",
		"create_table_cfsb_bindings",
		"create_table_cfsb_credentials",
		"create_table_rdpg_watch_notifications",
	}
	for _, key := range keys {
		k := strings.Replace(strings.Replace(key, "create_table_", "", 1), "_", ".", 1)
		sq := fmt.Sprintf(`SELECT table_name FROM information_schema.tables where table_schema='%s' AND table_name='%s';`, k[2], k[3])
		log.Trace(fmt.Sprintf("RDPG#initSchema() %s", sq))
		if err := db.QueryRow(sq).Scan(&name); err != nil {
			if err == sql.ErrNoRows {
				log.Trace(fmt.Sprintf("RDPG#initSchema() SQL[%s]", key))
				_, err = db.Exec(SQL[key])
				if err != nil {
					log.Error(fmt.Sprintf("RDPG#initSchema() %s", err))
				}
			} else {
				log.Error(fmt.Sprintf("rdpg.initSchema() %s", err))
				return err
			}
		}

	}

	// TODO: Move initial population of services out of rdpg-agent to Admin API.
	if err := db.QueryRow("SELECT name FROM cfsb.services WHERE name='rdpg' LIMIT 1;").Scan(&name); err != nil {
		if err == sql.ErrNoRows {
			if _, err = db.Exec(SQL["insert_default_cfsb_services"]); err != nil {
				log.Error(fmt.Sprintf("rdpg.initSchema(insert_default_cfsb_services) %s", err))
				return err
			}
		} else {
			log.Error(fmt.Sprintf("rdpg.initSchema() %s", err))
			return err
		}
	}

	// TODO: Move initial population of services out of rdpg-agent to Admin API.
	if err = db.QueryRow("SELECT name FROM cfsb.plans WHERE name='small' LIMIT 1;").Scan(&name); err != nil {
		if err == sql.ErrNoRows {
			if _, err = db.Exec(SQL["insert_default_cfsb_plans"]); err != nil {
				log.Error(fmt.Sprintf("rdpg.initSchema(insert_default_cfsb_plans) %s", err))
				return err
			}
		} else {
			log.Error(fmt.Sprintf("rdpg.initSchema() %s", err))
			return err
		}
	}

	return nil
}
