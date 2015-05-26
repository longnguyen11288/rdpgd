package main

import (
	"fmt"
	"os"
	"os/signal"
	"net/http"
	"io/ioutil"
	"github.com/gorilla/mux"
	"github.com/wayneeseguin/rdpg-agent/api"
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
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(1)
		}
	}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range ch {
			fmt.Printf("Received %v, shutting down...\n", sig)
			if err := os.Remove(pidFile) ; err != nil {
				fmt.Printf("%s\n",err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}()
	api.Run()
}
