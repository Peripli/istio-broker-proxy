package main

import (
	"flag"
	"fmt"
	"github.com/Peripli/istio-broker-proxy/pkg/config"
	m "github.com/Peripli/istio-broker-proxy/pkg/model"
	"github.com/Peripli/istio-broker-proxy/pkg/router"
	"istio.io/istio/pilot/pkg/model"
	"os"
)

func main() {
	var serviceName, endpointServiceEntry, hostVirtualService, systemDomain string
	var portServiceEntry int
	var clientConfig bool
	var help bool
	var delete bool

	flag.BoolVar(&clientConfig, "client", false, "Create client configuration")
	flag.StringVar(&serviceName, "service", "<service>", "name of the service")
	flag.StringVar(&hostVirtualService, "virtual-service", "<host>", "host of virtual service")
	flag.StringVar(&systemDomain, "system-domain", "<system-domain>", "system domain")
	flag.StringVar(&endpointServiceEntry, "endpoint", "<0.0.0.0>", "endpoint(ip) of the service entry")
	flag.IntVar(&portServiceEntry, "port", 99999, "port of the service entry")
	flag.BoolVar(&help, "help", false, "Print usage")
	flag.BoolVar(&delete, "delete", false, "Delete client config instead of creating")

	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	var configStore router.ConfigStore
	if clientConfig {
		configStore = router.NewExternKubeConfigStore("catalog")
	}

	createOutput(clientConfig, serviceName, hostVirtualService, portServiceEntry, endpointServiceEntry, systemDomain, delete, configStore)
}

func createOutput(clientConfig bool, serviceName string, hostVirtualService string, portServiceEntry int, endpointServiceEntry string, systemDomain string, delete bool, configStore router.ConfigStore) {
	var configs []model.Config
	if clientConfig {
		var err error
		id := fmt.Sprintf("client-binding-%s", serviceName)
		if delete {
			err = configStore.DeleteBinding(id)

		} else {
			_, err = router.CreateIstioObjectsInK8S(configStore, id, serviceName, m.Endpoint{Host: hostVirtualService, Port: 9000}, systemDomain)
		}
		if err != nil {
			fmt.Printf("error occured: %s", err.Error())
			os.Exit(1)
		}
	} else {
		configs = config.CreateEntriesForExternalService(serviceName, endpointServiceEntry, uint32(portServiceEntry), hostVirtualService, "client.my.client.domain.io", 9000, "")
		out, err := config.ToYamlDocuments(configs)
		if err == nil {
			fmt.Printf("%s", out)
		} else {
			fmt.Printf("error occured: %s", err.Error())
			os.Exit(1)
		}
	}
}
