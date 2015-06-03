package main

import (
	"fmt"
	"os"
	"syscall"
	"os/signal"
	"io/ioutil"
	"github.com/wayneeseguin/rdpg-agent/cfsb"
	"github.com/wayneeseguin/rdpg-agent/admin"
	"github.com/wayneeseguin/rdpg-agent/workers"
	"github.com/wayneeseguin/rdpg-agent/pg"
	"github.com/wayneeseguin/rdpg-agent/rdpg"
	"github.com/wayneeseguin/rdpg-agent/log"
)

var (
	pidFile string
)

func init() {
	pidFile = os.Getenv("RDPGAPI_PIDFILE")
}

func main() {
	if pidFile != "" {
		err := ioutil.WriteFile(pidFile,[]byte(string(os.Getpid())), 0644)
		if err != nil {
			log.Error(err.Error())
			os.Exit(1)
		}
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	go func() {
		for sig := range ch {
			log.Info(fmt.Sprintf("Received %v, shutting down...\n", sig))
			pg.Close()
			if _, err := os.Stat(pidFile); err == nil {
				if err := os.Remove(pidFile) ; err != nil {
					log.Error(err.Error())
					os.Exit(1)
				}
			}
			os.Exit(0)
		}
	}()

	err := pg.Open()
	if err != nil {
		log.Error(err.Error())
		proc, _ := os.FindProcess(os.Getpid())
		proc.Signal(syscall.SIGTERM)
	}

	rdpg.InitializeSchema()

	go cfsb.API()

	go admin.API()

	workers.Run()
}
