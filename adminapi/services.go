package adminapi

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/wayneeseguin/rdpgd/cfsbapi"
	"github.com/wayneeseguin/rdpgd/log"
	"github.com/wayneeseguin/rdpgd/rdpg"
)

type Service struct {
	Name string `db:"name" json:"name"`
}

func NewService(name string) (s Service, err error) {
	switch name {
	case "haproxy", "pgbouncer", "pgbdr", "consul":
		s = Service{Name: name}
	default:
	}
	return
}

func (s *Service) Configure() (err error) {
	log.Trace(fmt.Sprintf(`Service#Configure(%s)`, s.Name))

	switch s.Name {
	case "consul":
		return errors.New(`Service#Configure("consul") is not yet implemented`)
	case "haproxy":
		header, err := ioutil.ReadFile(`/var/vcap/jobs/rdpg/config/haproxy/haproxy.cfg.header`)
		if err != nil {
			log.Error(fmt.Sprintf("adminapi#Service.Configure(%s) ! %s", s.Name, err))
			return err
		}

		r := rdpg.NewRDPG()
		cluster, err := rdpg.NewCluster(r.ClusterID)
		if err != nil {
			log.Error(fmt.Sprintf(`cfsbapi.Instance#Configure() ! %s`, err))
		}
		node, err := cluster.WriteMaster()
		if err != nil {
			log.Error(fmt.Sprintf(`cfsbapi.Instance#Configure() ! %s`, err))
		}

		// TODO: 5432 & 6432 from environmental configuration.
		// TODO: Should this list come from active Consul registered hosts instead?
		footer := fmt.Sprintf(`
frontend pgbdr_write_port
bind 0.0.0.0:5432
  mode tcp
  default_backend pgbdr_write_master
 
backend pgbdr_write_master
  mode tcp
	server master %s:6432 check
	`, node.PG.IP)

		hc := []string{string(header), footer}
		err = ioutil.WriteFile(`/var/vcap/jobs/haproxy/config/haproxy.cfg`, []byte(strings.Join(hc, "\n")), 0640)
		if err != nil {
			log.Error(fmt.Sprintf("adminapi#Service.Configure(%s) ! %s", s.Name, err))
			return err
		}

		cmd := exec.Command("/var/vcap/jobs/haproxy/bin/control", "reload")
		err = cmd.Run()
		if err != nil {
			log.Error(fmt.Sprintf("adminapi#Service.Configure(%s) ! %s", s.Name, err))
			return err
		}

		return errors.New(`Service#Configure("haproxy") is not yet implemented`)
	case "pgbouncer":
		instances, err := cfsbapi.Instances()
		if err != nil {
			log.Error(fmt.Sprintf("adminapi#Service.Configure(%s) ! %s", s.Name, err))
			return err
		}

		pgbConf, err := ioutil.ReadFile(`/var/vcap/jobs/rdpg/config/pgbouncer/pgbouncer.ini`)
		if err != nil {
			log.Error(fmt.Sprintf("adminapi#Service.Configure(%s) ! %s", s.Name, err))
			return err
		}

		pgbUsers, err := ioutil.ReadFile(`/var/vcap/jobs/rdpg/config/pgbouncer/users`)
		if err != nil {
			log.Error(fmt.Sprintf("adminapi#Service.Configure(%s) ! %s", s.Name, err))
			return err
		}
		pc := []string{string(pgbConf)}
		pu := []string{string(pgbUsers)}
		for _, i := range instances {
			// TODO: Fetch port from something like os.Getenv("PG_PORT") instead of hardcoding here.
			c := fmt.Sprintf(`%s = host=%s port=%s dbname=%s`, i.Database, "127.0.0.1", "7432", i.Database)
			pc = append(pc, c)
			u := fmt.Sprintf(`"%s" "%s"`, i.User, i.Pass)
			pu = append(pu, u)
		}
		pc = append(pc, "")
		pu = append(pu, "")

		err = ioutil.WriteFile(`/var/vcap/store/pgbouncer/config/pgbouncer.ini`, []byte(strings.Join(pc, "\n")), 0640)
		if err != nil {
			log.Error(fmt.Sprintf("adminapi#Service.Configure(%s) ! %s", s.Name, err))
			return err
		}

		err = ioutil.WriteFile(`/var/vcap/store/pgbouncer/config/users`, []byte(strings.Join(pu, "\n")), 0640)
		if err != nil {
			log.Error(fmt.Sprintf("adminapi#Service.Configure(%s) ! %s", s.Name, err))
			return err
		}

		cmd := exec.Command("/var/vcap/jobs/pgbouncer/bin/control", "reload")
		err = cmd.Run()
		if err != nil {
			log.Error(fmt.Sprintf("adminapi#Service.Configure(%s) ! %s", s.Name, err))
			return err
		}
	case "pgbdr":
		// Add pg_hba.conf lines for the current datacenter cluster
		r := rdpg.NewRDPG()
		cluster, err := rdpg.NewCluster(r.ClusterID)
		if err != nil {
			// do something, yeah I'm tired
		}
		hbaHeader, err := ioutil.ReadFile(`/var/vcap/jobs/pgbdr/config/pg_hba.conf`)
		if err != nil {
			log.Error(fmt.Sprintf("adminapi#Service.Configure(%s) ! %s", s.Name, err))
			return err
		}
		hba := []string{string(hbaHeader)}

		for _, node := range cluster.Nodes {
			hba = append(hba, fmt.Sprintf(`host    replication   postgres %s/32  trust\n`, node.PG.IP))
			hba = append(hba, fmt.Sprintf(`host    all           postgres %s/32  trust\n`, node.PG.IP))
		}
		hba = append(hba, "")

		err = ioutil.WriteFile(`/var/vcap/store/pgbdr/data/pg_hba.conf`, []byte(strings.Join(hba, "\n")), 0640)
		if err != nil {
			log.Error(fmt.Sprintf("adminapi#Service.Configure(%s) ! %s", s.Name, err))
			return err
		}
		cmd := exec.Command("/var/vcap/jobs/pgbdr/bin/control", "reload")
		err = cmd.Run()
		if err != nil {
			log.Error(fmt.Sprintf("adminapi#Service.Configure(%s) ! %s", s.Name, err))
			return err
		}
	default:
		return errors.New(fmt.Sprintf(`Service#Configure("%s") is unknown.`, s.Name))
	}
	return
}
