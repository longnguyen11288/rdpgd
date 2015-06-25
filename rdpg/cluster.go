package rdpg

import (
	"fmt"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/wayneeseguin/rdpgd/log"
	"github.com/wayneeseguin/rdpgd/pg"
)

type Node struct {
	PG *pg.PG
}

type Cluster struct {
	Nodes []Node
	Role  string `json:"role" db:"role"`
	ID    string `json:"id" db:"cluster_id"`
}

func NewCluster(dc string) (c *Cluster, err error) {
	c = &Cluster{ID: dc}
	client, err := consulapi.NewClient(consulapi.DefaultConfig())
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.NewCluster() %s ! %s", dc, err))
		return
	}
	catalog := client.Catalog()
	q := consulapi.QueryOptions{Datacenter: dc}
	catalogNodes, _, err := catalog.Nodes(&q)
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.NewCluster() %s ! %s", dc, err))
		return
	}
	for _, catalogNode := range catalogNodes {
		node := Node{PG: &pg.PG{IP: catalogNode.Address}}
		c.Nodes = append(c.Nodes, node)
	}
	return
}

func (c *Cluster) SetWriteMaster(n *Node) (err error) {
	client, err := consulapi.NewClient(consulapi.DefaultConfig())
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.IAmWriteMaster() ! %s", err))
		return
	}
	agent := client.Agent()
	registration := &consulapi.AgentServiceRegistration{
		ID:      "postgresql",
		Name:    "master",
		Tags:    []string{},
		Address: n.PG.IP,
		Port:    5432,
		Check: &consulapi.AgentServiceCheck{
			Interval: "10s",
			Script:   "/var/vcap/jobs/rdpg-service/bin/check ha_pb_pg",
			TTL:      "0s",
			Timeout:  "5s",
		},
	}
	agent.ServiceRegister(registration)
	return
}

// Returns a cluster's write master Node
func (c *Cluster) WriteMaster() (n *Node, err error) {
	client, err := consulapi.NewClient(consulapi.DefaultConfig())
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.WriteMaster() ! %s", err))
		return
	}
	catalog := client.Catalog()
	q := consulapi.QueryOptions{Datacenter: c.ID}
	svc, _, err := catalog.Service("master", "", &q)
	if err != nil {
		log.Error(fmt.Sprintf(`rdpg.WriteMaster() ! %s`, err))
	}
	s := svc[0]
	n = &Node{PG: &pg.PG{IP: s.Address}}
	return
}
