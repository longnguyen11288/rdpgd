package config

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/wayneeseguin/rdpg-agent/services"
	"gopkg.in/yaml.v2"
)

/*
  TODO: Multiple CF's.
  syslog or logpath
*/

type CF struct {
	CCTarget  string `yaml:"cc_target"`
	UAATarget string `yaml:"uaa_target"`
	User      string `yaml:"user"`
	Password  string `yaml:"password"`
}

type Config struct {
	BrokerHost string `yaml:"broker_host"`
	BrokerPort string `yaml:"broker_port"`
	LogPath    string `yaml:"log_path"`
	CFs        []*CF  `yaml:"cfs"`
	Services   []*services.Service
}

var configPath string
var Conf Config

func init() {
	configPath = os.Getenv("RDPG_CONFIG")
	if configPath == "" {
		wd, _ := os.Getwd()
		configPath = path.Join(wd, "config.yml")
	}
}

func (c *Config) Load() (err error) {
	enc, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(enc, &Conf)
	if err != nil {
		return err
	}
	return nil
}
