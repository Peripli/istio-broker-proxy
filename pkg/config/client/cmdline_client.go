package main

import (
	"flag"
	"fmt"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/config"
	"os"
)

func main() {
	var serviceName, endpointServiceEntry, hostVirtualService string
	var portServiceEntry int

	flag.StringVar(&serviceName, "service", "<service>", "name of the service")
	flag.StringVar(&hostVirtualService, "virtual-service", "<host>", "host of virtual service")
	flag.StringVar(&endpointServiceEntry, "endpoint", "<0.0.0.0>", "endpoint(ip) of the service entry")
	flag.IntVar(&portServiceEntry, "port", 99999, "port of the service entry")

	flag.Parse()

	out, err := config.CreateEntriesForExternalService(serviceName, endpointServiceEntry, uint32(portServiceEntry), hostVirtualService)

	if err == nil {
		fmt.Printf("%s", out)
	} else {
		fmt.Printf("error occured: %s", err.Error())
		os.Exit(1)
	}
}
