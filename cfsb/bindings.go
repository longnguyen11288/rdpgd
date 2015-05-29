package cfsb

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
	URI      string
	UserName string
	Password string
}

