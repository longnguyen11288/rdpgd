package bdr

import (
	"fmt"
	"strings"

	"github.com/armon/consul-api"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/wayneeseguin/rdpgd/log"
	"github.com/wayneeseguin/rdpgd/pg"
	"github.com/wayneeseguin/rdpgd/rdpg"
)

type BDR struct {
	URI string
	DB  *sqlx.DB
}

func NewBDR(uri) (r *BDR) {
	r = &BDR{URI: uri}
	return
}

func (b *BDR) Hosts() (hosts []Host) {
	// TODO: Allow for managing multiple BDR clusters,
	//   for now we assume we are on the same cluster as the RDPG systems database.
	r := rdpg.NewRDPG()
	db, err := r.OpenDB("postgres")
	if err != nil {
		log.Error(fmt.Sprintf("bdr.BDR#Hosts() ! %s", err))
	}

	// TODO: Populate list of rdpg hosts for given URL,
	//`SELECT node_local_dsn FROM bdbdr_nodes INTO rdpg.hosts (node_local_dsn);`

	type dsn struct {
		DSN string `db:"node_local_dsn"`
	}

	dsns := []dsn{}
	err = db.Select(&dsns, SQL["bdr_nodes_dsn"])
	if err != nil {
		log.Error(fmt.Sprintf("RDPG#Hosts() %s ! %s", SQL["bdr_nodes"], err))
	}

	for _, t := range dsns {
		host := pg.Host{}
		s := strings.Split(t.DSN, " ")
		host.LocalDSN = t.DSN
		host.Host = strings.Split(s[0], "=")[1]
		host.Port = strings.Split(s[1], "=")[1]
		host.User = strings.Split(s[2], "=")[1]
		host.Database = `postgres` // strings.Split(s[3], "=")[1]
		hosts = append(hosts, host)
	}
	// TODO: Get this information into the database and then out of the rdpg.hosts
	//rows, err := db.Query("SELECT host,port,user,'postgres' FROM rdpg.hosts;")
	//if err != nil {
	//	log.Error(fmt.Sprintf("Hosts() %s", err))
	//} else {
	//	sqlx.StructScan(rows, hosts)
	//}
	db.Close()
	return hosts
}

// Question: Should we extract the BDR related functionality into a bd* package?
func (b *BDR) CreateUser(dbuser, dbpass string) (err error) {
	for _, host := range Hosts() {
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
	for _, host := range Hosts() {
		err = host.PGCreateDatabase(dbname, owner)
		if err != nil {
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
	for _, host := range Hosts() {
		err = host.PGCreateExtensions(dbname, exts)
		if err != nil {
			break
		}
	}
	return
}

func (b *BDR) CreateReplicationGroup(dbname string) (err error) {
	hosts := Hosts()
	// TODO: Drop Database on all hosts if err != nil for any operation below
	for index, host := range hosts {
		group := fmt.Sprintf("%s", host.IP)
		if index == 0 {
			err = host.PGBDRGroupCreate(group, dbname)
		} else {
			err = host.PGBDRGroupJoin(group, dbname, hosts[0])
		}
		if err != nil {
			break
		}
	}
	if err != nil {
		// Cleanup in BDR currently requires droping the database and trying again...
		DropDatabase(dbname)
		log.Error(fmt.Sprintf("bdr.BDR#CreateReplicationGroup(%s) ! %s", dbname, err))
	}
	return err
}

// Disable all usage of database.
func (b *BDR) DisableDatabase(dbname string) (err error) {
	hosts := Hosts()
	for i := len(hosts) - 1; i >= 0; i-- {
		err := hosts[i].PGDisableDatabase(dbname)
		if err != nil {
			log.Error(fmt.Sprintf("bdr.BDR#DisableDatabase(%s) %s ! %s", dbname, hosts[i].IP, err))
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
	hosts := Hosts()
	for i := len(hosts) - 1; i >= 0; i-- {
		hosts[i].Set(`database`, `postgres`)
		err = hosts[i].PGDropDatabase(dbname)
		if err != nil {
			log.Error(fmt.Sprintf("bdr.BDR#DropDatabase(%s) %s ! %s", dbname, hosts[i].IP, err))
		}
	}
	return nil
}

func (b *BDR) DropUser(dbuser string) (err error) {
	hosts := Hosts()
	for i := len(hosts) - 1; i >= 0; i-- {
		err = hosts[i].PGDropUser(dbuser)
		if err != nil {
			log.Error(fmt.Sprintf("RDPG#DropUser(%s) ! %s", dbuser, err))
		}
	}
	return nil
}

// Stop replication for given database (bdr replication group) and delete the grop on each node.
func (b *BDR) DeleteReplicationGroup(dbname string) (err error) {
	hosts := Hosts()
	for i := len(hosts) - 1; i >= 0; i-- {
		//hosts[i].Set(`database`, `postgres`)
		//db, err := hosts[i].Connect()
		//if err != nil {
		//	log.Error(fmt.Sprintf("RDPG#DropUser(%s) %s ! %s", dbname, hosts[i].IP, err))
		//	return err
		//}

		// TODO: Diable Replication for node...
		// Stop the replication
		//db.Close()
	}
	return nil
}

func IsWriteMaster() (b bool) {
	b = false
	client, _ := consulapi.NewClient(consulapi.DefaultConfig())
	catalog := client.Catalog()
	svc, _, err := catalog.Service("master", "", nil)
	if err != nil {
		log.Error(fmt.Sprintf(`rdpg.IsMaster() ! %s`, err))
	}
	// TODO: if the IP address matches our IP address we are master.
	if svc[0].Address == "" {

	}
	return
}
