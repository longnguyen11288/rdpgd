package rdpg

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/wayneeseguin/rdpg-agent/log"
)

var (
	rdpgURI string
)

func init() {
	rdpgURI = os.Getenv("RDPG_PG_URI")
}

func InitializeSchema() (err error) {
	log.Debug(fmt.Sprintf("rdpg.InitializeSchema() Connecting to DB URI: %s",rdpgURI))

	db, err := sqlx.Connect("postgres", rdpgURI)
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.InitializeSchema#Connect() %s:: %s\n", rdpgURI, err))
		return err
	}
	defer db.Close()
	
	log.Debug(fmt.Sprintf("rdpg.InitializeSchema() %s",SQLStatements["rdpg_extensions"]))
	if _, err = db.Exec(SQLStatements["rdpg_extensions"]); err != nil {
		log.Error(fmt.Sprintf("rdpg.InitializeSchema(rdpg_extensions) %s\n", err))
		return err
	}

	sq := "CREATE SCHEMA IF NOT EXISTS rdpg;"
	log.Debug(fmt.Sprintf("rdpg.InitializeSchema() %s",sq))
	if _, err = db.Exec(sq); err != nil {
		log.Error(fmt.Sprintf("rdpg.InitializeSchema(create_table_services) %s\n", err))
		return err
	}

	log.Debug(fmt.Sprintf("rdpg.InitializeSchema() %s",SQLStatements["create_table_rdpg_services"]))
	if _, err = db.Exec(SQLStatements["create_table_rdpg_services"]); err != nil {
		log.Error(fmt.Sprintf("rdpg.InitializeSchema(create_table_rdpg_services) %s\n", err))
		return err
	}

	log.Debug(fmt.Sprintf("rdpg.InitializeSchema() %s",SQLStatements["create_table_rdpg_plans"]))
	if _, err = db.Exec(SQLStatements["create_table_plans"]); err != nil {
		log.Error(fmt.Sprintf("rdpg.InitializeSchema(create_table_plans) %s\n", err))
		return err
	}

	type Service struct {
		Name string `db:"name"`
	}
	s := &Service{}

	if err = db.Get(&s, "SELECT name FROM rdpg.services WHERE name='rdpg' LIMIT 1;"); err != nil {
		log.Error(fmt.Sprintf("rdpg.InitializeSchema() %s\n", err))
		return err
	}
	if s.Name != "rdpg" {
		if _, err = db.Exec(SQLStatements["insert_default_rdpg_services"]); err != nil {
			log.Error(fmt.Sprintf("rdpg.InitializeSchema(insert_default_rdpg_services) %s\n", err))
			return err
		}
	}

	if err = db.Get(&s, "SELECT name FROM rdpg.plans WHERE name='small' LIMIT 1;"); err != nil {
		log.Error(fmt.Sprintf("rdpg.InitializeSchema() %s\n", err))
		return err
	}
	if s.Name != "small" {
		if _, err = db.Exec(SQLStatements["insert_default_rdpg_plans"]); err != nil {
			log.Error(fmt.Sprintf("rdpg.InitializeSchema(insert_default_rdpg_plans) %s\n", err))
			return err
		}
	}

	return nil
}
