package rdpg

import (
	"fmt"

	_ "github.com/lib/pq"
	"github.com/wayneeseguin/rdpg-agent/log"
)

func (r *RDPG) CreateUser(username, password string) (err error) {
	for _, host := range r.Hosts() {
		host.Database = `postgres`
		// TODO: Switch to using host.CreateUser() after allowing to specify a list of roles (SUPERUSER, etc...)
		db, err := host.Connect()
		if err != nil {
			log.Error(fmt.Sprintf(`RDPG#CreateUser(%s) %s ! %s`, username, host.Host, err))
			return err
		}

		// TODO: Check if user exists first
		sq := fmt.Sprintf(`CREATE USER %s;`, username)
		log.Trace(fmt.Sprintf(`RDPG#CreateUser(%s) %s > %s`, username, host.Host, sq))
		_, err = db.Exec(sq)
		if err != nil {
			log.Error(fmt.Sprintf("RDPG#CreateUser(%s) %s ! %s", username, host.Host, err))
			db.Close()
			return err
		}

		sq = fmt.Sprintf(`ALTER USER %s ENCRYPTED PASSWORD '%s'`, username, password)
		log.Trace(fmt.Sprintf(`RDPG#CreateUser(%s) %s > %s`, username, host.Host, sq))
		_, err = db.Exec(sq)
		if err != nil {
			log.Error(fmt.Sprintf("RDPG#CreateUser(%s) %s ! %s", username, host.Host, err))
		}
		db.Close()
	}
	return nil
}

func (r *RDPG) CreateDatabase(dbname, owner string) (err error) {
	for _, host := range r.Hosts() {
		err = host.CreateDatabase(dbname, owner)
		if err != nil {
			break
		}

		host.Database = dbname
		err = host.CreateExtensions([]string{"btree_gist", "bdr"})
		if err != nil {
			break
		}

	}
	if err != nil {
		// Cleanup in BDR currently requires droping the database and trying again...
		r.DropDatabase(dbname)
	}
	return err
}

func (r *RDPG) CreateReplicationGroup(dbname string) (err error) {
	hosts := r.Hosts()
	// TODO: Drop Database on all hosts if err != nil for any operation below
	for index, host := range hosts {
		host.Database = dbname
		db, err := host.Connect()
		if err != nil {
			break
		}
		sq := ""
		name := fmt.Sprintf("%s", host.Host)
		if index == 0 {
			sq = fmt.Sprintf(`SELECT bdr.bdr_group_create(
				local_node_name := '%s',
				node_external_dsn := 'host=%s port=%s user=%s dbname=%s'
			); `, name, host.Host, host.Port, host.User, dbname)
		} else {
			sq = fmt.Sprintf(`SELECT bdr.bdr_group_join(
				local_node_name := '%s',
				node_external_dsn := 'host=%s port=%s user=%s dbname=%s',
				join_using_dsn := 'host=%s port=%s user=%s dbname=%s'
			); `,
				name, host.Host, host.Port, host.User, host.Database,
				hosts[0].Host, hosts[0].Port, hosts[0].User, dbname,
			)
		}
		log.Trace(fmt.Sprintf(`RDPG#CreateReplicationGroup(%s) %s > %s`, dbname, host.Host, sq))
		_, err = db.Exec(sq)
		if err == nil {
			sq = `SELECT bdr.bdr_node_join_wait_for_ready();`
			log.Trace(fmt.Sprintf(`RDPG#CreateReplicationGroup(%s) %s > %s`, dbname, host.Host, sq))
			_, err = db.Exec(sq)
		}
		db.Close()
	}
	if err != nil {
		// Cleanup in BDR currently requires droping the database and trying again...
		log.Error(fmt.Sprintf("CreateReplicationGroup(%s) ! %s", dbname, err))
	}
	return err
}

func (r *RDPG) DisableDatabase(dbname string) (err error) {
	hosts := r.Hosts()
	for i := len(hosts) - 1; i >= 0; i-- {
		host := hosts[i]

		// TODO: move this into host.DisableDatabase()
		host.Database = "postgres"
		db, err := host.Connect()
		if err != nil {
			log.Error(fmt.Sprintf("RDPG#DisableDatabase(%s) %s ! %s", dbname, host.Host, err))
			return err
		}
		sq := fmt.Sprintf(`SELECT rdpg.bdr_disable_database('%s');`, dbname)
		log.Trace(fmt.Sprintf(`RDPG#DisableDatabase(%s) DISABLE %s > %s`, dbname, host.Host, sq))
		_, err = db.Exec(sq)
		if err != nil {
			log.Error(fmt.Sprintf("RDPG#DisableDatabase(%s) DISABLE %s ! %s", dbname, host.Host, err))
		}
		db.Close()
	}

	return nil
}

func (r *RDPG) BackupDatabase(dbname string) (err error) {
	log.Error(fmt.Sprintf("RDPG#BackupDatabase(%s) TODO: IMPLEMENT", dbname))
	return nil
}

func (r *RDPG) DropDatabase(dbname string) (err error) {
	hosts := r.Hosts()
	for i := len(hosts) - 1; i >= 0; i-- {
		host := hosts[i]

		host.Database = "postgres"
		// TODO: move this into host.DropDatabase()
		db, err := host.Connect()
		if err != nil {
			log.Error(fmt.Sprintf("RDPG#DropDatabase(%s) %s ! %s", dbname, host.Host, err))
			return err
		}

		// sq := fmt.Sprintf(SELECT slot_name FROM pg_replication_slots WHERE database='%s',dbname);
		// pg_recvlogical --drop-slot

		// TODO: How do we drop a database in bdr properly?
		sq := fmt.Sprintf(`DROP DATABASE IF EXISTS %s`, dbname)
		log.Trace(fmt.Sprintf(`RDPG#DropDatabase(%s) %s DROP > %s`, dbname, host.Host, sq))
		_, err = db.Exec(sq)
		if err != nil {
			log.Error(fmt.Sprintf("RDPG#DropDatabase(%s) DROP %s ! %s", dbname, host.Host, err))
		}
		db.Close()
	}
	return nil
}

func (r *RDPG) DropUser(name string) (err error) {
	hosts := r.Hosts()
	for i := len(hosts) - 1; i >= 0; i-- {
		host := hosts[i]

		host.Database = "postgres"
		db, err := host.Connect()
		if err != nil {
			log.Error(fmt.Sprintf("RDPG#DropUser(%s) %s ! %s", name, host.Host, err))
			return err
		}

		// TODO: move this into host.DropUser()
		sq := fmt.Sprintf(`DROP USER %s`, name)
		log.Trace(fmt.Sprintf(`RDPG#DropUser(%s) %s > %s`, name, host.Host, sq))
		_, err = db.Exec(sq)
		if err != nil {
			log.Error(fmt.Sprintf("RDPG#DropUser(%s) %s ! %s", name, host.Host, err))
		}
		db.Close()
	}
	return nil
}
