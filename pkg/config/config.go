package config

import (
	"github.com/ghodss/yaml"
	"istio.io/istio/pilot/pkg/config/kube/crd"
	"istio.io/istio/pilot/pkg/model"
)

type generatedServiceConfig struct {
	gateway        model.Config
	virtualService model.Config
	serviceEntry   model.Config
}

func CreateEntriesForExternalService(serviceName string, endpointServiceEntry string, portServiceEntry uint32, hostVirtualService string) (string, error) {

	var entry generatedServiceConfig

	entry.gateway = createIngressGatewayForExternalService(hostVirtualService, 9000, serviceName, "client.istio.sapcloud.io")
	entry.virtualService = createIngressVirtualServiceForExternalService(hostVirtualService, portServiceEntry, serviceName)
	entry.serviceEntry = createServiceEntryForExternalService(endpointServiceEntry, portServiceEntry, serviceName)

	return toYamlArray(entry)
}

func CreateEntriesForExternalServiceClient(serviceName string, hostName string, portNumber uint32) (string, error) {
	var array []interface{}

	gateway := createEgressGatewayForExternalService(hostName, 443, serviceName)
	kubernetesConf, _ := crd.ConvertConfig(model.Gateway, gateway)
	array = append(array, kubernetesConf)

	destinationRule := createEgressDestinationRuleForExternalService(hostName, 9000, serviceName)
	kubernetesConf, _ = crd.ConvertConfig(model.DestinationRule, destinationRule)
	array = append(array, kubernetesConf)
	array = append(array, kubernetesConf)
	array = append(array, kubernetesConf)

	serviceEntry := createEgressInternServiceEntryForExternalService(hostName, portNumber, serviceName)
	kubernetesConf, _ = crd.ConvertConfig(model.ServiceEntry, serviceEntry)
	array = append(array, kubernetesConf)

	serviceEntry = createEgressExternServiceEntryForExternalService(hostName, 9000, serviceName)
	kubernetesConf, _ = crd.ConvertConfig(model.ServiceEntry, serviceEntry)
	array = append(array, kubernetesConf)

	bytes, err := yaml.Marshal(array)
	return string(bytes), err
}

func toYamlArray(entry generatedServiceConfig) (string, error) {
	var array []interface{}

	array = addConfig(array, model.Gateway, entry.gateway)
	array = addConfig(array, model.ServiceEntry, entry.serviceEntry)
	array = addConfig(array, model.VirtualService, entry.virtualService)

	bytes, err := yaml.Marshal(array)
	return string(bytes), err
}

func addConfig(array []interface{}, schema model.ProtoSchema, config model.Config) []interface{} {
	kubernetesConf, err := crd.ConvertConfig(schema, config)
	if err == nil {
		array = append(array, kubernetesConf)
	}

	return array
}
