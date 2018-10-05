package config

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/profiles"
	"istio.io/istio/pilot/pkg/config/kube/crd"
	istioModel "istio.io/istio/pilot/pkg/model"
	"regexp"
	"strings"
)

const (
	gateway         = "Gateway"
	serviceEntry    = "ServiceEntry"
	virtualService  = "VirtualService"
	destinationRule = "DestinationRule"
)

var invalidIdentifiers = regexp.MustCompile(`[^0-9a-z-]`)

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

func CreateIstioConfigForProvider(request *model.BindRequest, response *model.BindResponse, bindingId string, systemDomain string) []istioModel.Config {
	var istioConfig []istioModel.Config
	for index, endpoint := range response.Endpoints {
		portServiceEntry := uint32(endpoint.Port)
		consumerId := request.NetworkData.Data.ConsumerId
		ingressPort := uint32(9000)

		serviceName := createValidIdentifer(fmt.Sprintf("%d-%s", index, bindingId))
		endpointServiceEntry := endpoint.Host
		hostVirtualService := profiles.CreateEndpointHosts(bindingId, systemDomain, index)
		istioConfig = append(istioConfig,
			CreateEntriesForExternalService(serviceName, endpointServiceEntry, portServiceEntry, hostVirtualService, consumerId, ingressPort)...)
	}
	return istioConfig
}

func createValidIdentifer(identifer string) string {
	validIdentifier := invalidIdentifiers.ReplaceAllString(strings.ToLower(identifer), "-")
	if strings.HasPrefix(validIdentifier, "-") {
		validIdentifier = strings.TrimPrefix(validIdentifier, "-")
	}
	return validIdentifier

}

func CreateEntriesForExternalServiceClient(serviceName string, hostName string, serviceIP string, port int) []istioModel.Config {
	var configs []istioModel.Config

	serviceEntry := createEgressExternServiceEntryForExternalService(hostName, uint32(port), serviceName)
	configs = append(configs, serviceEntry)

	virtualService := createMeshVirtualServiceForExternalService(hostName, 443, serviceName, serviceIP)
	configs = append(configs, virtualService)
	virtualService = createEgressVirtualServiceForExternalService(hostName, uint32(port), serviceName, 443)
	configs = append(configs, virtualService)

	gateway := createEgressGatewayForExternalService(hostName, 443, serviceName)
	configs = append(configs, gateway)

	destinationRule := createEgressDestinationRuleForExternalService(hostName, uint32(port), serviceName)
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
	//kubernetesConf, err := ToRuntimeObject(config)
	//if err != nil {
	//	return "", err
	//}
	bytes, err := yaml.Marshal(config)
	return string(bytes), err
}

func ToRuntimeObject(config istioModel.Config) (crd.IstioObject, error) {
	schema := schemas[config.Type]
	return crd.ConvertConfig(schema, config)
}
