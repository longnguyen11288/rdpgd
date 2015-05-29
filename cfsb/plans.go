package cfsb

type Metadata struct {
	Cost        string              `json:"cost"`
	Bullets     []map[string]string `json:"bullets"`
	DisplayName string              `json:"displayname"`
}

type Plan struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Metadata    Metadata `json:"metadata"`
	MgmtDbUri   string   `json:""`
}

