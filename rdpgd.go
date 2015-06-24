package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/wayneeseguin/rdpgd/adminapi"
	"github.com/wayneeseguin/rdpgd/cfsbapi"
	"github.com/wayneeseguin/rdpgd/log"
	"github.com/wayneeseguin/rdpgd/rdpg"
	"github.com/wayneeseguin/rdpgd/tasks"
)

var (
	VERSION string
	pidFile string
	Role    string
)

func init() {
	pidFile = os.Getenv("RDPGD_PIDFILE")
}

// TODO: Allow for --version
func main() {
	go signalHandler()

	writePidFile()

	parseArgs()

	switch Role {
	case "manager":
		manager()
	case "service":
		service()
	case "bootstrap":
		bootstrap()
	}
}

func parseArgs() {
	for index, arg := range os.Args {
		if index == 0 {
			continue
		}

		switch arg {
		case "bootstrap":
			Role = "bootstrap"
		case "manager":
			Role = "manager"
		case "service":
			Role = "service"
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
	rdpg

	Usage:

	rdpg [Flag(s)] <Action>

	Actions:

	bootstrap Bootstrap RDPG schemas, filesystem etc...
	version   print rdpg version
	help      print this message

	Flags:

	--version  print rdpg version
	--help     print this message

	`)
}

func manager() (err error) {
	log.Info(`Starting with 'manager' role...`)
	bootstrap()
	go cfsbapi.Listen()
	go tasks.Scheduler(Role)
	go tasks.Work(Role)
	adminapi.Listen()
	return
}

func service() (err error) {
	log.Info(`Starting with 'service' role...`)
	bootstrap()
	go tasks.Scheduler(Role)
	go tasks.Work(Role)
	adminapi.Listen()
	return
}

func bootstrap() (err error) {
	r := rdpg.NewRDPG()
	err = r.Bootstrap(Role)
	if err != nil {
		log.Error(fmt.Sprintf(`Bootstrap(%s) failed`, Role))
		proc, _ := os.FindProcess(os.Getpid())
		proc.Signal(syscall.SIGTERM)
	}
	return
}

func writePidFile() {
	if pidFile != "" {
		pid := os.Getpid()
		log.Trace(fmt.Sprintf(`main.writePidFile() Writing pid %d to %s`, pid, pidFile))
		err := ioutil.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
		if err != nil {
			log.Error(fmt.Sprintf(`main.writePidFile() Error while writing pid '%d' to '%s' :: %s`, pid, pidFile, err))
			os.Exit(1)
		}
	}
	return
}

func signalHandler() (err error) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	for sig := range ch {
		log.Info(fmt.Sprintf("main.signalHandler() Received signal %v, shutting down gracefully...", sig))
		if _, err := os.Stat(pidFile); err == nil {
			if err := os.Remove(pidFile); err != nil {
				log.Error(err.Error())
				os.Exit(1)
			}
		}
		os.Exit(0)
	}
	return
}
