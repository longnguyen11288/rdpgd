package rdpg

import (
	"fmt"

	"github.com/armon/consul-api"
	"github.com/coreos/go-log/log"
)

func IsMaster() (b bool) {
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
