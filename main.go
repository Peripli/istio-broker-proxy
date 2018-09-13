package main

import (
	"flag"
	"fmt"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/router"
	"log"
	"os"
	"strconv"
)

const (
	DefaultPort = 8080
)

var port int = DefaultPort

func readPort() {
	portAsString := os.Getenv("PORT")
	if len(portAsString) != 0 {
		var err error
		port, err = strconv.Atoi(portAsString)
		if nil != err {
			port = DefaultPort
		}
	}
}

func main() {
	flag.IntVar(&port, "port", DefaultPort, "port to be used")
	router.SetupConfiguration()
	flag.Parse()
	readPort()

	log.Printf("Running on port %d\n", port)

	router := router.SetupRouter()
	router.Run(fmt.Sprintf(":%d", port))
}
