package cfsbapi

import (
	"fmt"
	"strings"

	"github.com/wayneeseguin/rdpgd/log"
	"github.com/wayneeseguin/rdpgd/rdpg"
)

type Credentials struct {
	URI      string `json:"uri"`
	DSN      string `json:"dsn"`
	JDBCURI  string `json:"jdbc_uri"`
	Host     string `db:"host" json:"host"`
	Port     string `db:"port" json:"port"`
	UserName string `db:"username" json:"username"`
	Password string `db:"password" json:"password"`
	Database string `db:"database" json:"database"`
}

type Binding struct {
	Id         int         `db:"id"`
	BindingId  string      `json:"binding_id"`
	InstanceId string      `json:"instance_id"`
	Creds      Credentials `json:"credentials"`
}

func CreateBinding(instanceId, bindingId string) (binding *Binding, err error) {
	instance, err := FindInstance(instanceId)
	if err != nil {
		log.Error(fmt.Sprintf(`cfsbapi.CreateBinding(%s,%s) ! %s`, instanceId, bindingId, err))
		return
	}
	binding = &Binding{BindingId: bindingId, InstanceId: instanceId}

	dns := instance.ExternalDNS()
	s := strings.Split(dns, ":")

	binding.Creds = Credentials{
		URI:      instance.URI(),
		DSN:      instance.DSN(),
		JDBCURI:  "jdbc:" + instance.URI(),
		Host:     s[0],
		Port:     s[1],
		UserName: instance.User,
		Password: instance.Pass,
		Database: instance.Database,
	}

	r := rdpg.NewRDPG()
	err = r.OpenDB("rdpg")
	if err != nil {
		log.Error(fmt.Sprintf("cfsbapi#CreateBinding(%s) ! %s", bindingId, err))
	}
	defer r.DB.Close()

	sq := `INSERT INTO cfsbapi.bindings (instance_id,binding_id) VALUES (?,?);`
	_, err = r.DB.Query(sq, binding.InstanceId, binding.BindingId)
	if err != nil {
		log.Error(fmt.Sprintf(`cfsbapi.CreateBinding(%s) ! %s`, bindingId, err))
	}

	sq = `INSERT INTO cfsbapi.credentials (instance_id,binding_id,host,port,dbuser,dbpass,dbname) VALUES (?,?,?,?,?,?,?);`
	_, err = r.DB.Query(sq, binding.InstanceId, binding.BindingId, binding.Creds.Host, binding.Creds.Port, binding.Creds.UserName, binding.Creds.Password, binding.Creds.Database)
	if err != nil {
		log.Error(fmt.Sprintf(`cfsbapi.CreateBinding(%s) ! %s`, bindingId, err))
	}

	return
}

func RemoveBinding(bindingId string) (binding *Binding, err error) {
	binding, err = FindBinding(bindingId)
	if err != nil {
		log.Error(fmt.Sprintf(`cfsbapi.CreateBinding(%s) ! %s`, bindingId, err))
		return
	}
	r := rdpg.NewRDPG()
	sq := `UPDATE cfsbapi.bindings SET ineffective_at = CURRENT_TIMESTAMP WHERE binding_id = ?;`
	log.Trace(fmt.Sprintf(`cfsbapi.RemoveBinding(%s) > %s`, bindingId, sq))
	err = r.OpenDB("rdpg")
	if err != nil {
		log.Error(fmt.Sprintf("cfsbapi#RemoveBinding(%s) ! %s", bindingId, err))
	}
	defer r.DB.Close()

	_, err = r.DB.Query(sq, bindingId)
	if err != nil {
		log.Error(fmt.Sprintf(`cfsbapi.CreateBinding(%s) ! %s`, bindingId, err))
	}
	return
}

func FindBinding(bindingId string) (binding *Binding, err error) {
	r := rdpg.NewRDPG()
	b := Binding{}
	sq := `SELECT id,instance_id, binding_id FROM cfsbapi.bindings WHERE binding_id=lower(?) LIMIT 1;`
	log.Trace(fmt.Sprintf(`cfsbapi.FindBinding(%s) > %s`, bindingId, sq))
	err = r.OpenDB("rdpg")
	if err != nil {
		log.Error(fmt.Sprintf("cfsbapi#FindBinding(%s) ! %s", bindingId, err))
	}
	defer r.DB.Close()

	err = r.DB.Get(&b, sq, bindingId)
	if err != nil {
		// TODO: Change messaging if err is sql.NoRows then say couldn't find binding with bindingId
		log.Error(fmt.Sprintf("cfsbapi.FindBinding(%s) ! %s", bindingId, err))
	}
	binding = &b
	return
}
