package main

import (
	"flag"
	"fmt"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/router"
	"log"
	"os"
)

var producerConfig router.ProducerConfig
var consumerConfig router.ConsumerConfig
var routerConfig router.RouterConfig

func main() {
	SetupConfiguration()
	flag.Parse()

	log.Printf("Running on port %d\n", routerConfig.Port)
	var interceptor router.ServiceBrokerInterceptor
	if consumerConfig.ConsumerId != "" {
		interceptor = router.NewConsumerInterceptor(consumerConfig)
	} else if producerConfig.ProviderId != "" {
		interceptor = router.NewProducerInterceptor(producerConfig, routerConfig.Port)
	} else {
		interceptor = router.NoOpInterceptor{}
	}

	engine := router.SetupRouter(interceptor, routerConfig)
	engine.Run(fmt.Sprintf(":%d", routerConfig.Port))
}

func SetupConfiguration() {
	flag.StringVar(&producerConfig.SystemDomain, "systemdomain", "", "system domain of the landscape")
	flag.StringVar(&producerConfig.ProviderId, "providerId", "", "The subject alternative name of the provider for which the service has a certificate")

	flag.IntVar(&producerConfig.LoadBalancerPort, "loadBalancerPort", 0, "port of the load balancer of the landscape")
	flag.StringVar(&producerConfig.IstioDirectory, "istioDirectory", os.TempDir(), "Directory to store the istio configuration files")
	flag.StringVar(&producerConfig.IpAddress, "ipAddress", "127.0.0.1", "IP address of ingress")

	flag.StringVar(&consumerConfig.ConsumerId, "consumerId", "", "The subject alternative name of the consumer for which the service has a certificate")

	flag.StringVar(&routerConfig.ForwardURL, "forwardUrl", "", "url for forwarding incoming requests")
	flag.IntVar(&routerConfig.Port, "port", router.DefaultPort, "Server listen port")
}
