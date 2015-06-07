package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/wayneeseguin/rdpg-agent/admin"
	"github.com/wayneeseguin/rdpg-agent/cfsb"
	"github.com/wayneeseguin/rdpg-agent/log"
	"github.com/wayneeseguin/rdpg-agent/rdpg"
	"github.com/wayneeseguin/rdpg-agent/workers"
)

var (
	pidFile string
)

func init() {
	pidFile = os.Getenv("RDPG_AGENT_PIDFILE")
}

func main() {
	if pidFile != "" {
		err := ioutil.WriteFile(pidFile, []byte(string(os.Getpid())), 0644)
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
			if _, err := os.Stat(pidFile); err == nil {
				if err := os.Remove(pidFile); err != nil {
					log.Error(err.Error())
					os.Exit(1)
				}
			}
			os.Exit(0)
		}
	}()

	r := rdpg.New()
	err := r.OpenDB()
	if err != nil {
		log.Error(err.Error())
		proc, _ := os.FindProcess(os.Getpid())
		proc.Signal(syscall.SIGTERM)
	}

	go cfsb.API()

	go admin.API()

	workers.Run()
}
