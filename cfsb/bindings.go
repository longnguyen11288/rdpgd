package cfsb

import (
	"fmt"
	"strings"

	"github.com/wayneeseguin/rdpg-agent/log"
	"github.com/wayneeseguin/rdpg-agent/rdpg"
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
	Id         string      `json:"binding_id"`
	InstanceId string      `json:"instance_id"`
	Creds      Credentials `json:"credentials"`
}

func CreateBinding(instanceId, bindingId string) (binding *Binding, err error) {
	instance, err := FindInstance(instanceId)
	if err != nil {
		return
	}
	binding = &Binding{Id: bindingId, InstanceId: instanceId}

	dns := instance.ExternalDNS()
	s := strings.Split(dns, ":")

	binding.Creds = Credentials{
		URI:      instance.URI(),
		DSN:      instance.DSN(),
		JDBCURI:  "jdbc:" + instance.URI(),
		Host:     s[0],
		Port:     s[1],
		UserName: instance.Database,
		Password: instance.Pass,
		Database: instance.User,
	}

	r := rdpg.New()

	sq := `INSERT INTO cfsb.bindings (instance_id,binding_id) VALUES ($1,$2);`
	_, err = r.DB.Query(sq, binding.InstanceId, binding.Id)
	if err != nil {
		log.Error(fmt.Sprintf(`cfsb.CreateBinding() %s\n`, err))
	}

	sq = `INSERT INTO cfsb.credentials (instance_id,binding_id,host,port,uname,pass,dbname) VALUES ($1,$2,$3,$4,$5,$6,$7);`
	_, err = r.DB.Query(sq, binding.InstanceId, binding.Id, binding.Creds.Host, binding.Creds.Port, binding.Creds.UserName)
	if err != nil {
		log.Error(fmt.Sprintf(`cfsb.CreateBinding() %s\n`, err))
	}

	return
}
