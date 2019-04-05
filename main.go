package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/Peripli/istio-broker-proxy/pkg/router"
	"log"
	"net/url"
)

var producerInterceptor router.ProducerInterceptor
var consumerInterceptor router.ConsumerInterceptor
var routerConfig router.Config
var serviceNamePrefix string
var networkProfile string
var configStore string


func newConfigStore(configStoreURL string) (router.ConfigStore, error) {
	uri , err := url.Parse(configStoreURL)
	if err != nil {
		return nil, err
	}
	switch uri.Scheme {
	case "k8s":return router.NewInClusterConfigStore(), nil
	case "file": return router.NewFileConfigStore(uri.Path), nil
	default:
		return nil, errors.New("Invalid schema for config store:" + uri.Scheme)
	}
}

func newConfigStoreOrFail(configStoreURL string) router.ConfigStore {
	store, err := newConfigStore(configStoreURL)
	if err != nil {
		panic(err)
	}
	return store
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	SetupConfiguration()
	flag.Parse()
	engine := router.SetupRouter(configureInterceptor(newConfigStoreOrFail), routerConfig)
	engine.Run(fmt.Sprintf(":%d", routerConfig.Port))
}

func configureInterceptor(configStoreFactory func(configStoreUrl string) router.ConfigStore) router.ServiceBrokerInterceptor {
	if networkProfile == "" {
		panic("networkProfile not configured")
	}
	log.Printf("Running on port %d\n", routerConfig.Port)
	var interceptor router.ServiceBrokerInterceptor
	if consumerInterceptor.ConsumerID != "" {
		consumerInterceptor.ServiceNamePrefix = serviceNamePrefix
		consumerInterceptor.NetworkProfile = networkProfile
		consumerInterceptor.ConfigStore = configStoreFactory(configStore)
		interceptor = consumerInterceptor
	} else if producerInterceptor.ProviderID != "" {
		producerInterceptor.ServiceNamePrefix = serviceNamePrefix
		producerInterceptor.NetworkProfile = networkProfile
		producerInterceptor.ConfigStore = configStoreFactory(configStore)
		err := producerInterceptor.WriteIstioConfigFiles(routerConfig.Port)
		if err != nil {
			panic(fmt.Sprintf("unable to write istio-broker provider side configuration file: %v", err))
		}
		interceptor = producerInterceptor
	} else {
		interceptor = router.NewNoOpInterceptor()
	}
	return interceptor
}

// SetupConfiguration sets up the configuration (e.g. initializing the available command line parameters)
func SetupConfiguration() {
	flag.StringVar(&producerInterceptor.SystemDomain, "systemdomain", "", "system domain of the landscape")
	flag.StringVar(&producerInterceptor.ProviderID, "providerId", "", "The subject alternative name of the provider for which the service has a certificate")

	flag.IntVar(&producerInterceptor.LoadBalancerPort, "loadBalancerPort", 9000, "port of the load balancer of the landscape")
	flag.StringVar(&configStore, "configStore", "k8s://", "URL to store the istio configuration files. Use 'k8s://' to store the configuration to kubernetes")
	flag.StringVar(&producerInterceptor.IPAddress, "ipAddress", "127.0.0.1", "IP address of ingress")
	flag.StringVar(&producerInterceptor.PlanMetaData, "planMetaData", "{}", "Metadata which is added to each service")
	flag.StringVar(&networkProfile, "networkProfile", "", "Network profile e.g. urn:local.test:public")

	flag.StringVar(&consumerInterceptor.ConsumerID, "consumerId", "", "The subject alternative name of the consumer for which the service has a certificate")

	flag.StringVar(&routerConfig.ForwardURL, "forwardUrl", "", "url for forwarding incoming requests")
	flag.BoolVar(&routerConfig.SkipVerifyTLS, "skipVerifyTLS", false, "Do not verify the certificate of the forwardUrl")
	flag.IntVar(&routerConfig.Port, "port", router.DefaultPort, "Server listen port")
	flag.StringVar(&serviceNamePrefix, "serviceNamePrefix", "", "Service name prefix")
}
