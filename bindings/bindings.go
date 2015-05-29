package bindings

import (
	"github.com/jmoiron/sqlx"
	//"github.com/wayneeseguin/rdpg-agent/pg"
)

type Credentials struct {
	Uri      string `json:"uri"`
	UserName string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Database string `json:"database"`
}

type ServiceBinding struct {
	Id       string
	DBConn   *sqlx.DB
	UserName string
	Password string
}

