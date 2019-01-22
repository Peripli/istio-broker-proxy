package config

import (
	"github.com/gogo/protobuf/proto"
	"istio.io/istio/pilot/pkg/model"
)

func createServiceEntryForExternalService(endpointAddress string, portNumber uint32, serviceName string) model.Config {

	return wrapAsConfig(createRawServiceEntryForExternalService(endpointAddress, portNumber, serviceName),
		serviceEntry, serviceName+"-service-entry")
}

func createIngressVirtualServiceForExternalService(hostName string, port uint32, serviceName string) model.Config {
	return wrapAsConfig(createRawIngressVirtualServiceForExternalService(hostName, port, serviceName),
		virtualService, serviceName+"-virtual-service")
}

func createIngressGatewayForExternalService(hostName string, portNumber uint32, serviceName string, clientName string) model.Config {
	return wrapAsConfig(createRawIngressGatewayForExternalService(hostName, portNumber, clientName),
		gateway, serviceName+"-gateway")
}

func wrapAsConfig(spec proto.Message, typeName string, name string) model.Config {
	config := model.Config{Spec: spec}
	config.Type = typeName
	config.Name = name
	return config
}
