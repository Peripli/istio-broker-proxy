package main

import (
	"flag"
	"fmt"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/config"
	"istio.io/istio/pilot/pkg/model"
	"os"
)

func main() {
	var serviceName, endpointServiceEntry, hostVirtualService, serviceIp string
	var portServiceEntry int
	var clientConfig bool
	var help bool

	flag.BoolVar(&clientConfig, "client", false, "Create client configuration")
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

	createOutput(clientConfig, serviceName, hostVirtualService, portServiceEntry, endpointServiceEntry, serviceIp)
}

func createOutput(clientConfig bool, serviceName string, hostVirtualService string, portServiceEntry int, endpointServiceEntry string, serviceIp string) {
	var configs []model.Config
	if clientConfig {
		configs = config.CreateEntriesForExternalServiceClient(serviceName, hostVirtualService, serviceIp, 9000)
	} else {
		configs = config.CreateEntriesForExternalService(serviceName, endpointServiceEntry, uint32(portServiceEntry), hostVirtualService, "client.istio.sapcloud.io", 9000)
	}
	out, err := config.ToYamlDocuments(configs)
	if err == nil {
		fmt.Printf("%s", out)
	} else {
		fmt.Printf("error occured: %s", err.Error())
		os.Exit(1)
	}
}
