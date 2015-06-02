package rdpg

import(
	"fmt"
	"github.com/wayneeseguin/rdpg-agent/log"
)

func InitializeSchema() (err error) {
	n := NewNode("127.0.0.1", "5432", "postgres", "rdpg")
	db,err := n.Connect()
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.InitializeSchema(): %s\n", err))
		return err
	}
	defer db.Close()

	if _,err = db.Exec(SQLStatements["create_table_services"]) ; err != nil {
		log.Error(fmt.Sprintf("rdpg.InitializeSchema(create_table_services): %s\n", err))
		return err
	}
	if _,err = db.Exec(SQLStatements["create_table_plans"]) ; err != nil {
		log.Error(fmt.Sprintf("rdpg.InitializeSchema(create_table_plans): %s\n", err))
		return err
	}

	var name string

	if err = db.Get(&name, "SELECT name FROM rdpg.services WHERE name='rdpg' LIMIT 1;") ; err != nil {
		return err
	}
	if name != "rdpg" {
		if _,err = db.Exec(SQLStatements["insert_default_services"]) ; err != nil {
			log.Error(fmt.Sprintf("rdpg.InitializeSchema(insert_default_services): %s\n", err))
			return err
		}
	}

	if err = db.Get(&name, "SELECT name FROM rdpg.plans WHERE name='small' LIMIT 1;") ; err != nil  {
		return err
	}
	if name != "small" {
		if _,err = db.Exec(SQLStatements["insert_default_plans"]) ; err != nil {
			log.Error(fmt.Sprintf("rdpg.InitializeSchema(insert_default_plans): %s\n", err))
			return err
		}
	}

	return nil
}

