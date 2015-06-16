package cfsb

import (
	"fmt"

	"github.com/wayneeseguin/rdpg-agent/log"
	"github.com/wayneeseguin/rdpg-agent/rdpg"
)

type PlanDetails struct {
	Cost        string              `json:"cost"`
	Bullets     []map[string]string `json:"bullets"`
	DisplayName string              `json:"displayname"`
}

type Plan struct {
	Id          string      `db:"id"`
	PlanId      string      `db:"plan_id" json:"id"`
	Name        string      `db:"name" json:"name"`
	Description string      `db:"description" json:"description"`
	Metadata    PlanDetails `json:"metadata"`
	MgmtDbUri   string      `json:""`
}

func FindPlan(planId string) (plan *Plan, err error) {
	r := rdpg.New()
	r.OpenDB("rdpg")
	plan = &Plan{}
	sq := `SELECT id,name,description FROM cfsb.plans WHERE id=$1 LIMIT 1;`
	err = r.DB.Get(&plan, sq, planId)
	if err != nil {
		log.Error(fmt.Sprintf("cfsb.FindPlan(%s) %s", planId, err))
	}
	r.DB.Close()
	return plan, err
}
