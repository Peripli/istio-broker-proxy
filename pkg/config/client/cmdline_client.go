package main

import (
	"flag"
	"fmt"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/config"
	"os"
)

func main() {
	var serviceName, endpointServiceEntry, hostVirtualService, serviceIp string
	var portServiceEntry int
	var help bool

	flag.StringVar(&serviceName, "service", "<service>", "name of the service")
	flag.StringVar(&hostVirtualService, "virtual-service", "<host>", "host of virtual service")
	flag.StringVar(&endpointServiceEntry, "endpoint", "<0.0.0.0>", "endpoint(ip) of the service entry")
	flag.StringVar(&serviceIp, "service-ip", "<0.0.0.0>", "consumer side service ip")
	flag.IntVar(&portServiceEntry, "port", 99999, "port of the service entry")
	flag.BoolVar(&help, "help", false, "Print usage")

	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	createOutput(serviceName, hostVirtualService, portServiceEntry, endpointServiceEntry, serviceIp)
}

func createOutput(serviceName string, hostVirtualService string, portServiceEntry int, endpointServiceEntry string, serviceIp string) {
	configs := config.CreateEntriesForExternalService(serviceName, endpointServiceEntry, uint32(portServiceEntry), hostVirtualService, "client.istio.sapcloud.io", 9000)
	out, err := config.ToYamlDocuments(configs)
	if err == nil {
		fmt.Printf("%s", out)
	} else {
		fmt.Printf("error occured: %s", err.Error())
		os.Exit(1)
	}
}
