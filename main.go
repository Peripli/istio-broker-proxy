package main

import (
	"flag"
	"fmt"
	"github.com/Peripli/istio-broker-proxy/pkg/router"
	"log"
	"os"
)

var producerInterceptor router.ProducerInterceptor
var consumerInterceptor router.ConsumerInterceptor
var routerConfig router.RouterConfig
var serviceNamePrefix string
var networkProfile string

func main() {
	SetupConfiguration()
	flag.Parse()
	engine := router.SetupRouter(configureInterceptor(router.NewInClusterConfigStore), routerConfig)
	engine.Run(fmt.Sprintf(":%d", routerConfig.Port))
}

func configureInterceptor(configStoreFactory func() router.ConfigStore) router.ServiceBrokerInterceptor {
	if networkProfile == "" {
		panic("networkProfile not configured")
	}
	log.Printf("Running on port %d\n", routerConfig.Port)
	var interceptor router.ServiceBrokerInterceptor
	if consumerInterceptor.ConsumerId != "" {
		consumerInterceptor.ServiceNamePrefix = serviceNamePrefix
		consumerInterceptor.NetworkProfile = networkProfile
		consumerInterceptor.ConfigStore = configStoreFactory()
		interceptor = consumerInterceptor
	} else if producerInterceptor.ProviderId != "" {
		producerInterceptor.ServiceNamePrefix = serviceNamePrefix
		producerInterceptor.NetworkProfile = networkProfile
		err := producerInterceptor.WriteIstioConfigFiles(routerConfig.Port)
		if err != nil {
			panic(fmt.Sprintf("Unable to write istio-broker provider side configuration file: %v", err))
		}
		interceptor = producerInterceptor
	} else {
		interceptor = router.NoOpInterceptor{}
	}
	return interceptor
}

// SetupConfiguration sets up the configuration (e.g. initializing the available command line parameters)
func SetupConfiguration() {
	flag.StringVar(&producerInterceptor.SystemDomain, "systemdomain", "", "system domain of the landscape")
	flag.StringVar(&producerInterceptor.ProviderId, "providerId", "", "The subject alternative name of the provider for which the service has a certificate")

	flag.IntVar(&producerInterceptor.LoadBalancerPort, "loadBalancerPort", 9000, "port of the load balancer of the landscape")
	flag.StringVar(&producerInterceptor.IstioDirectory, "istioDirectory", os.TempDir(), "Directory to store the istio configuration files")
	flag.StringVar(&producerInterceptor.IpAddress, "ipAddress", "127.0.0.1", "IP address of ingress")
	flag.StringVar(&producerInterceptor.PlanMetaData, "planMetaData", "{}", "Metadata which is added to each service")
	flag.StringVar(&networkProfile, "networkProfile", "", "Network profile e.g. urn:local.test:public")

	flag.StringVar(&consumerInterceptor.ConsumerId, "consumerId", "", "The subject alternative name of the consumer for which the service has a certificate")

	flag.StringVar(&routerConfig.ForwardURL, "forwardUrl", "", "url for forwarding incoming requests")
	flag.BoolVar(&routerConfig.SkipVerifyTLS, "skipVerifyTLS", false, "Do not verify the certificate of the forwardUrl")
	flag.IntVar(&routerConfig.Port, "port", router.DefaultPort, "Server listen port")
	flag.StringVar(&serviceNamePrefix, "serviceNamePrefix", "", "Service name prefix")
}
