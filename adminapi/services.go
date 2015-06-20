package adminapi

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/wayneeseguin/rdpg-agent/cfsbapi"
	"github.com/wayneeseguin/rdpg-agent/log"
	"github.com/wayneeseguin/rdpg-agent/rdpg"
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
		header, err := ioutil.ReadFile(`/var/vcap/jobs/rdpg-agent/config/haproxy/haproxy.cfg.header`)
		if err != nil {
			log.Error(fmt.Sprintf("cfsbapi#Service.Configure(%s) ! %s", s.Name, err))
			return err
		}

		r := rdpg.New()
		hosts := r.Hosts()
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
	`, hosts[0].Host)

		hc := []string{string(header), footer}
		err = ioutil.WriteFile(`/var/vcap/jobs/haproxy/config/haproxy.cfg`, []byte(strings.Join(hc, "\n")), 0640)
		if err != nil {
			log.Error(fmt.Sprintf("cfsbapi#Service.Configure(%s) ! %s", s.Name, err))
			return err
		}

		cmd := exec.Command("/var/vcap/jobs/haproxy/bin/control", "reload")
		err = cmd.Run()
		if err != nil {
			log.Error(fmt.Sprintf("cfsbapi#Service.Configure(%s) ! %s", s.Name, err))
			return err
		}

		return errors.New(`Service#Configure("haproxy") is not yet implemented`)
	case "pgbouncer":
		instances, err := cfsbapi.Instances()
		if err != nil {
			log.Error(fmt.Sprintf("cfsbapi#Service.Configure(%s) ! %s", s.Name, err))
			return err
		}

		pgbConf, err := ioutil.ReadFile(`/var/vcap/jobs/rdpg-agent/config/pgbouncer/pgbouncer.ini`)
		if err != nil {
			log.Error(fmt.Sprintf("cfsbapi#Service.Configure(%s) ! %s", s.Name, err))
			return err
		}

		pgbUsers, err := ioutil.ReadFile(`/var/vcap/jobs/rdpg-agent/config/pgbouncer/users`)
		if err != nil {
			log.Error(fmt.Sprintf("cfsbapi#Service.Configure(%s) ! %s", s.Name, err))
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
			log.Error(fmt.Sprintf("cfsbapi#Service.Configure(%s) ! %s", s.Name, err))
			return err
		}

		err = ioutil.WriteFile(`/var/vcap/store/pgbouncer/config/users`, []byte(strings.Join(pu, "\n")), 0640)
		if err != nil {
			log.Error(fmt.Sprintf("cfsbapi#Service.Configure(%s) ! %s", s.Name, err))
			return err
		}

		cmd := exec.Command("/var/vcap/jobs/pgbouncer/bin/control", "reload")
		err = cmd.Run()
		if err != nil {
			log.Error(fmt.Sprintf("cfsbapi#Service.Configure(%s) ! %s", s.Name, err))
			return err
		}
	case "pgbdr":
		return errors.New(`Service#Configure("pgbdr") is not yet implemented`)
	default:
		return errors.New(fmt.Sprintf(`Service#Configure("%s") is unknown.`, s.Name))
	}
	return
}
