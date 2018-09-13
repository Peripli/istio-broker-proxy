package config

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
	"istio.io/istio/pilot/pkg/config/kube/crd"
	istioModel "istio.io/istio/pilot/pkg/model"
)

const (
	gateway         = "Gateway"
	serviceEntry    = "ServiceEntry"
	virtualService  = "VirtualService"
	destinationRule = "DestinationRule"
)

var schemas = map[string]istioModel.ProtoSchema{
	gateway:         istioModel.Gateway,
	serviceEntry:    istioModel.ServiceEntry,
	virtualService:  istioModel.VirtualService,
	destinationRule: istioModel.DestinationRule,
}

func CreateEntriesForExternalService(serviceName string, endpointServiceEntry string, portServiceEntry uint32, hostVirtualService string, clientName string, ingressPort uint32) []istioModel.Config {
	var configs []istioModel.Config

	configs = append(configs, createIngressGatewayForExternalService(hostVirtualService, ingressPort, serviceName, clientName))
	configs = append(configs, createIngressVirtualServiceForExternalService(hostVirtualService, portServiceEntry, serviceName))
	configs = append(configs, createServiceEntryForExternalService(endpointServiceEntry, portServiceEntry, serviceName))

	return configs
}

func CreateIstioConfigForProvider(request *model.BindRequest, response *model.BindResponse, bindingId string) []istioModel.Config {
	var istioConfig []istioModel.Config
	for _, endpoint := range response.Endpoints {
		originalEndpointHost := endpoint.Host
		portServiceEntry := uint32(endpoint.Port)
		ingressDomain := "services.cf.dev01.aws.istio.sapcloud.io"
		consumerId := request.NetworkData.Data.ConsumerId
		ingressPort := uint32(9000)
		serviceName := fmt.Sprintf("%s-%v", bindingId, originalEndpointHost)
		endpointServiceEntry := originalEndpointHost
		hostVirtualService := fmt.Sprintf("%s-%v-%s", bindingId, originalEndpointHost, ingressDomain)
		istioConfig = append(istioConfig,
			CreateEntriesForExternalService(serviceName, endpointServiceEntry, portServiceEntry, hostVirtualService, consumerId, ingressPort)...)
	}
	return istioConfig
}

func CreateEntriesForExternalServiceClient(serviceName string, hostName string, portNumber uint32) []istioModel.Config {
	var configs []istioModel.Config

	serviceEntry := createEgressInternServiceEntryForExternalService(hostName, portNumber, serviceName)
	configs = append(configs, serviceEntry)

	serviceEntry = createEgressExternServiceEntryForExternalService(hostName, 9000, serviceName)
	configs = append(configs, serviceEntry)

	virtualService := createMeshVirtualServiceForExternalService(hostName, 443, serviceName, portNumber)
	configs = append(configs, virtualService)
	virtualService = createEgressVirtualServiceForExternalService(hostName, 9000, serviceName, 443)
	configs = append(configs, virtualService)

	gateway := createEgressGatewayForExternalService(hostName, 443, serviceName)
	configs = append(configs, gateway)

	destinationRule := createEgressDestinationRuleForExternalService(hostName, 9000, serviceName)
	configs = append(configs, destinationRule)

	destinationRule = createSidecarDestinationRuleForExternalService(hostName, serviceName)
	configs = append(configs, destinationRule)

	return configs
}

func ToYamlDocuments(entry []istioModel.Config) (string, error) {
	var result, text string
	var err error

	for _, element := range entry {
		text, err = toText(element)
		result += "---\n" + text
	}

	return result, err
}

func toText(config istioModel.Config) (string, error) {
	schema := schemas[config.Type]
	kubernetesConf, err := crd.ConvertConfig(schema, config)
	if err != nil {
		return "", err
	}
	bytes, err := yaml.Marshal(kubernetesConf)
	return string(bytes), err
}
