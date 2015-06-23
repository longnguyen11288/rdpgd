package bdr

import (
	"fmt"
	"strings"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/wayneeseguin/rdpgd/log"
	"github.com/wayneeseguin/rdpgd/pg"
	"github.com/wayneeseguin/rdpgd/rdpg"
)

var SQL map[string]string = map[string]string{
	"bdr_nodes":     `SELECT * FROM bdr.bdr_nodes;`,
	"bdr_nodes_dsn": `SELECT node_local_dsn FROM bdr.bdr_nodes;`,
}

type BDR struct {
	URI string
	DB  *sqlx.DB
}

func NewBDR(uri) (r *BDR) {
	r = &BDR{URI: uri}
	return
}

func (b *BDR) Nodes() (nodes []Node) {
	// TODO: Allow for managing multiple BDR clusters, the list of nodes should not
	// be coming from the nodes themselves but instead through the configuration
	// they were registered from.
	//   for now we assume we are on the same cluster as the RDPG systems database.

	r := rdpg.NewRDPG()
	err := r.OpenDB("postgres")
	if err != nil {
		log.Error(fmt.Sprintf("bdr.BDR#Nodes() ! %s", err))
	}
	defer r.DB.Close()

	type dsn struct {
		DSN string `db:"node_local_dsn"`
	}

	dsns := []dsn{}
	err = r.DB.Select(&dsns, SQL["bdr_nodes_dsn"])
	if err != nil {
		log.Error(fmt.Sprintf("bdr.BDR#Nodes() %s ! %s", SQL["bdr_nodes"], err))
	}

	for _, t := range dsns {
		node := pg.PG{}
		s := strings.Split(t.DSN, " ")
		node.IP = strings.Split(s[0], "=")[1]
		node.Port = strings.Split(s[1], "=")[1]
		node.User = strings.Split(s[2], "=")[1]
		node.Database = `postgres` // strings.Split(s[3], "=")[1]
		nodes = append(nodes, node)
	}
	return nodes
}

// Question: Should we extract the BDR related functionality into a bd* package?
func (b *BDR) CreateUser(dbuser, dbpass string) (err error) {
	for _, host := range Nodes() {
		host.Set(`database`, `postgres`)
		err = host.PGCreateUser(dbuser, dbpass)
		if err != nil {
			log.Error(fmt.Sprintf(`bdr.BDR#CreateUser(%s) %s ! %s`, dbuser, host.IP, err))
			return err
		}
	}
	return nil
}

func (b *BDR) CreateDatabase(dbname, owner string) (err error) {
	for _, host := range Nodes() {
		err = host.PGCreateDatabase(dbname, owner)
		if err != nil {
			log.Error(fmt.Sprintf(`bdr.BDR<%s>#CreateDatabase(%s,%s) %s ! %s`, host.IP, dbname, owner, err))
			break
		}
	}
	if err != nil {
		// Cleanup in BDR currently requires droping the database and trying again...
		DropDatabase(dbname)
	}
	return
}

func (b *BDR) CreateExtensions(dbname string, exts []string) (err error) {
	for _, host := range Nodes() {
		err = host.PGCreateExtensions(dbname, exts)
		if err != nil {
			log.Error(fmt.Sprintf(`bdr.BDR<%s>#CreateExtensions(%s) %s ! %s`, host.IP, dbname, ext, err))
			break
		}
	}
	return
}

func (b *BDR) CreateReplicationGroup(dbname string) (err error) {
	hosts := Nodes()
	// TODO: Drop Database on all hosts if err != nil for any operation below
	for index, host := range hosts {
		group := fmt.Sprintf("%s", host.IP)
		if index == 0 {
			err = host.PGBDRGroupCreate(group, dbname)
		} else {
			err = host.PGBDRGroupJoin(group, dbname, hosts[0])
		}
		if err != nil {
			log.Error(fmt.Sprintf(`bdr.BDR<%s>#CreateExtensions(%s) ! %s`, host.IP, dbname, err))
			break
		}
	}
	if err != nil {
		// Cleanup in BDR currently requires droping the database and trying again...
		DropDatabase(dbname)
	}
	return err
}

// Disable all usage of database.
func (b *BDR) DisableDatabase(dbname string) (err error) {
	hosts := Nodes()
	for i := len(hosts) - 1; i >= 0; i-- {
		err := hosts[i].PGDisableDatabase(dbname)
		if err != nil {
			log.Error(fmt.Sprintf("bdr.BDR<%s>#DisableDatabase(%s) ! %s", hosts[i].IP, dbname, err))
			return err
		}
	}
	return nil
}

func (b *BDR) BackupDatabase(dbname string) (err error) {
	log.Error(fmt.Sprintf("bdr.BDR#BackupDatabase(%s) TODO: IMPLEMENT", dbname))
	return nil
}

func (b *BDR) DropDatabase(dbname string) (err error) {
	hosts := Nodes()
	for i := len(hosts) - 1; i >= 0; i-- {
		hosts[i].Set(`database`, `postgres`)
		err = hosts[i].PGDropDatabase(dbname)
		if err != nil {
			log.Error(fmt.Sprintf("bdr.BDR<%s>#DropDatabase(%s) ! %s", hosts[i].IP, dbname, err))
		}
	}
	return nil
}

func (b *BDR) DropUser(dbuser string) (err error) {
	hosts := Nodes()
	for i := len(hosts) - 1; i >= 0; i-- {
		err = hosts[i].PGDropUser(dbuser)
		if err != nil {
			log.Error(fmt.Sprintf("bdr.BDR<%s>#DropUser(%s) ! %s", hosts[i].IP, dbuser, err))
		}
	}
	return nil
}

// Stop replication for given database (bdr replication group) and delete the grop on each node.
func (b *BDR) DeleteReplicationGroup(dbname string) (err error) {
	hosts := Nodes()
	for i := len(hosts) - 1; i >= 0; i-- {
		//hosts[i].Set(`database`, `postgres`)
		//db, err := hosts[i].Connect()
		//if err != nil {
		//	log.Error(fmt.Sprintf("bdr.BDR<%s>#DeleteReplicationGroup(%s) ! %s", hosts[i].IP, dbname, err))
		//	return err
		//}

		// TODO: Diable Replication for node...
		// Stop the replication
		//db.Close()
	}
	return nil
}

func IAmWriteMaster() (b bool) {
	b = false

	client, err := consulapi.NewClient(consulapi.DefaultConfig())
	if err != nil {
		log.Error(fmt.Sprintf("rdpg.IAmWriteMaster() ! %s", err))
		return
	}
	agent := client.Agent()
	info, err := agent.Self()

	dc := info["Config"]["Datacenter"].(string)
	myIP := info["Config"]["AdvertiseAddr"].(string)

	catalog := client.Catalog()
	q := consulapi.QueryOptions{Datacenter: dc}
	svc, _, err := catalog.Service("master", "", &q)
	if err != nil {
		log.Error(fmt.Sprintf(`rdpg.IAmWriteMaster() ! %s`, err))
	}

	if svc[0].Address == myIP {
		b = true
	}

	return
}
