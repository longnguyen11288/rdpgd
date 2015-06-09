package admin

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/wayneeseguin/rdpg-agent/cfsb"
	"github.com/wayneeseguin/rdpg-agent/log"
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
	case "haproxy":
		return errors.New(`Service#Configure("haproxy") is not yet implemented`)
	case "pgbouncer":
		pgbConf, err := ioutil.ReadFile(`/var/vcap/jobs/pgbouncer/config/pgbouncer.ini`)
		if err != nil {
			return err
		}

		pgbUsers, err := ioutil.ReadFile(`/var/vcap/jobs/pgbouncer/config/users`)
		if err != nil {
			return err
		}

		instances, err := cfsb.Instances()
		if err != nil {
			return err
		}

		pc := []string{string(pgbConf)}
		pu := []string{string(pgbUsers)}
		for _, i := range instances {
			// TODO: Fetch port from config somewhere instead of hardcoding here.
			c := fmt.Sprintf(`%s = host=%s port=%s dbname=%s`, i.Database, "127.0.0.1", "7432", i.Database)
			pc = append(pc, c)
			u := fmt.Sprintf(`"%s" "%s"`, i.User, i.Pass)
			pu = append(pu, u)
		}

		err = ioutil.WriteFile(`/var/vcap/store/pgbouncer/config/pgbouncer.ini`, []byte(strings.Join(pc, "\n")), 0640)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(`/var/vcap/store/pgbouncer/config/users`, []byte(strings.Join(pu, "\n")), 0640)
		if err != nil {
			return err
		}

		cmd := exec.Command("/var/vcap/jobs/pgbouncer/bin/pgbouncer", "reload")
		err = cmd.Run()
		if err != nil {
			return err
		}
	case "pgbdr":
		return errors.New(`Service#Configure("pgbdr") is not yet implemented`)
	case "consul":
		return errors.New(`Service#Configure("consul") is not yet implemented`)
	default:
		return errors.New(fmt.Sprintf(`Service#Configure("%s") is unknown.`, s.Name))
	}
	return
}
