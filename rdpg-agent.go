package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/wayneeseguin/rdpg-agent/adminapi"
	"github.com/wayneeseguin/rdpg-agent/cfsbapi"
	"github.com/wayneeseguin/rdpg-agent/log"
	"github.com/wayneeseguin/rdpg-agent/rdpg"
	"github.com/wayneeseguin/rdpg-agent/scheduler"
	"github.com/wayneeseguin/rdpg-agent/workers"
)

var (
	VERSION string
	pidFile string
)

func init() {
	pidFile = os.Getenv("RDPG_AGENT_PIDFILE")
	ParseArgs()
}

// TODO: Allow for --version
func main() {
	if pidFile != "" {
		pid := os.Getpid()
		log.Debug(fmt.Sprintf("Writing pid %d to %s", pid, pidFile))
		err := ioutil.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
		if err != nil {
			log.Error(fmt.Sprintf(`Error while writing pid '%d' to '%s' :: %s`, pid, pidFile, err))
			os.Exit(1)
		}
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	go func() {
		for sig := range ch {
			log.Info(fmt.Sprintf("Received %v, shutting down...", sig))
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
	err := r.OpenDB("rdpg")
	if err != nil {
		log.Error(err.Error())
		proc, _ := os.FindProcess(os.Getpid())
		proc.Signal(syscall.SIGTERM)
	}

	go cfsbapi.Listen()

	go scheduler.Schedule()

	go workers.Work()

	adminapi.Listen()
}

func ParseArgs() {
	for index, arg := range os.Args {
		if index == 0 {
			continue
		}
		switch arg {
		case "init":
			r := rdpg.New()
			err := r.OpenDB("rdpg")
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}

			err = r.InitSchema()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		case "version", "--version", "-version":
			fmt.Fprintf(os.Stdout, "%s\n", VERSION)
			os.Exit(0)
		case "help", "-h", "?", "--help":
			usage()
			os.Exit(0)
		default:
			usage()
			os.Exit(1)
		}
	}
}

func usage() {
	fmt.Fprintf(os.Stdout, `
rdpg-agent

Usage:

  rdpg-agent [Flag(s)] <Action>

Actions:

  init     Initialize rdpg database schema
  version  print rdpg-agent version
  help     print this message

Flags:

  --version  print rdpg-agent version
  --help     print this message

`)
}
