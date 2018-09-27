package main

import (
	"flag"
	"fmt"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/router"
	"log"
	"os"
)

func main() {
	SetupConfiguration()
	flag.Parse()

	log.Printf("Running on port %d\n", router.ProxyConfiguration.Port)

	engine := router.SetupRouter()
	engine.Run(fmt.Sprintf(":%d", router.ProxyConfiguration.Port))
}

func SetupConfiguration() {
	flag.StringVar(&router.ProxyConfiguration.ForwardURL, "forwardUrl", "", "url for forwarding incoming requests")
	flag.StringVar(&router.ProxyConfiguration.SystemDomain, "systemdomain", "", "system domain of the landscape")
	flag.StringVar(&router.ProxyConfiguration.ProviderId, "providerId", "", "The subject alternative name of the provider for which the service has a certificate")
	flag.StringVar(&router.ProxyConfiguration.ConsumerId, "consumerId", "", "The subject alternative name of the consumer for which the service has a certificate")
	flag.IntVar(&router.ProxyConfiguration.LoadBalancerPort, "loadBalancerPort", 0, "port of the load balancer of the landscape")
	flag.StringVar(&router.ProxyConfiguration.IstioDirectory, "istioDirectory", os.TempDir(), "Directory to store the istio configuration files")
	flag.StringVar(&router.ProxyConfiguration.IpAddress, "ipAddress", "127.0.0.1", "IP address of ingress")
	flag.IntVar(&router.ProxyConfiguration.Port, "port", router.DefaultPort, "Server listen port")
}
