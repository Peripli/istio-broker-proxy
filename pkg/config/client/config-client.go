package main

import (
	"flag"
	"fmt"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/config"
	"os"
)

func main() {
	var serviceName, endpointServiceEntry, hostVirtualService string
	var portServiceEntry uint

	flag.StringVar(&serviceName, "service", "", "name of the service")
	flag.StringVar(&endpointServiceEntry, "endpoint-service-entry", "", "ip of the service entry")
	flag.UintVar(&portServiceEntry, "port-service-entry", 0, "port of the service entry")
	flag.StringVar(&hostVirtualService, "host-virtual-service", "", "host of the virtual service")
	flag.Parse()

	if (serviceName == "") || (endpointServiceEntry == "") || (portServiceEntry == 0) || (hostVirtualService == "") {
		flag.PrintDefaults()
		os.Exit(1)
	}

	out, err := config.CreateEntriesForExternalService(serviceName, endpointServiceEntry, uint32(portServiceEntry), hostVirtualService)

	if err == nil {
		fmt.Printf("%s", out)
	} else {
		fmt.Printf("error occured: %s", err.Error())
		os.Exit(1)
	}

}
