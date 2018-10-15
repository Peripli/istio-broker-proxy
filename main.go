package main

import (
	"flag"
	"fmt"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/router"
	"log"
	"os"
)

var producerInterceptor router.ProducerInterceptor
var consumerInterceptor router.ConsumerInterceptor
var routerConfig router.RouterConfig
var systemDomain string

func main() {
	SetupConfiguration()
	flag.Parse()

	log.Printf("Running on port %d\n", routerConfig.Port)
	var interceptor router.ServiceBrokerInterceptor
	if consumerInterceptor.ConsumerId != "" {
		if len(systemDomain) == 0 {
			consumerInterceptor.SystemDomain = "cluster.local"
		} else {
			consumerInterceptor.SystemDomain = systemDomain
		}
		consumerInterceptor.ConfigStore = router.NewInClusterConfigStore()
		interceptor = consumerInterceptor
	} else if producerInterceptor.ProviderId != "" {
		producerInterceptor.SystemDomain = systemDomain
		err := producerInterceptor.WriteIstioConfigFiles(routerConfig.Port)
		if err != nil {
			panic(fmt.Sprintf("Unable to write istio-broker provider side configuration file: %v", err))
		}
		interceptor = producerInterceptor
	} else {
		interceptor = router.NoOpInterceptor{}
	}

	engine := router.SetupRouter(interceptor, routerConfig)
	engine.Run(fmt.Sprintf(":%d", routerConfig.Port))
}

func SetupConfiguration() {
	flag.StringVar(&systemDomain, "systemdomain", "", "system domain of the landscape")
	flag.StringVar(&producerInterceptor.ProviderId, "providerId", "", "The subject alternative name of the provider for which the service has a certificate")

	flag.IntVar(&producerInterceptor.LoadBalancerPort, "loadBalancerPort", 9000, "port of the load balancer of the landscape")
	flag.StringVar(&producerInterceptor.IstioDirectory, "istioDirectory", os.TempDir(), "Directory to store the istio configuration files")
	flag.StringVar(&producerInterceptor.IpAddress, "ipAddress", "127.0.0.1", "IP address of ingress")

	flag.StringVar(&consumerInterceptor.ConsumerId, "consumerId", "", "The subject alternative name of the consumer for which the service has a certificate")
	flag.StringVar(&consumerInterceptor.Namespace, "namespace", "default", "Kubernetes consumer side namespace")

	flag.StringVar(&routerConfig.ForwardURL, "forwardUrl", "", "url for forwarding incoming requests")
	flag.IntVar(&routerConfig.Port, "port", router.DefaultPort, "Server listen port")
}
