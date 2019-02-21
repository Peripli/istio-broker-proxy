package config

import (
	"fmt"
	"github.com/Peripli/istio-broker-proxy/pkg/model"
	"github.com/Peripli/istio-broker-proxy/pkg/profiles"
	"github.com/ghodss/yaml"
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
const (
	istio_gateway         = "gateway"
	istio_serviceEntry    = "service-entry"
	istio_virtualService  = "virtual-service"
	istio_destinationRule = "destination-rule"
)

var invalidIdentifiers = regexp.MustCompile(`[^0-9a-z-]`)

var schemas = map[string]istioModel.ProtoSchema{
	gateway:               istioModel.Gateway,
	serviceEntry:          istioModel.ServiceEntry,
	virtualService:        istioModel.VirtualService,
	destinationRule:       istioModel.DestinationRule,
	istio_gateway:         istioModel.Gateway,
	istio_serviceEntry:    istioModel.ServiceEntry,
	istio_virtualService:  istioModel.VirtualService,
	istio_destinationRule: istioModel.DestinationRule,
}

type ServiceId struct {
	Type string
	Name string
}

func CreateEntriesForExternalService(serviceName string, endpointServiceEntry string, portServiceEntry uint32, hostVirtualService string, clientName string, ingressPort uint32, providerSAN string) []istioModel.Config {
	var configs []istioModel.Config

	configs = append(configs, createIngressGatewayForExternalService(hostVirtualService, ingressPort, serviceName, clientName, providerSAN))
	configs = append(configs, createIngressVirtualServiceForExternalService(hostVirtualService, portServiceEntry, serviceName))
	configs = append(configs, createServiceEntryForExternalService(endpointServiceEntry, portServiceEntry, serviceName))

	return configs
}

func CreateIstioConfigForProvider(request *model.BindRequest, response *model.BindResponse, bindingId string, systemDomain string, providerSAN string) []istioModel.Config {
	var istioConfig []istioModel.Config
	for index, endpoint := range response.Endpoints {
		portServiceEntry := uint32(endpoint.Port)
		consumerId := request.NetworkData.Data.ConsumerId
		ingressPort := uint32(9000)

		serviceName := createValidIdentifer(fmt.Sprintf("%d-%s", index, bindingId))
		endpointServiceEntry := endpoint.Host
		hostVirtualService := profiles.CreateEndpointHosts(bindingId, systemDomain, index)
		istioConfig = append(istioConfig,
			CreateEntriesForExternalService(serviceName, endpointServiceEntry, portServiceEntry, hostVirtualService, consumerId, ingressPort, providerSAN)...)
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

func DeleteEntriesForExternalServiceClient(serviceName string) []ServiceId {
	result := make([]ServiceId, 0)
	result = append(result, sidecarDestinationRuleForExternalService(serviceName))
	result = append(result, egressDestinationRuleForExternalService(serviceName))
	result = append(result, egressGatewayForExternalService(serviceName))
	result = append(result, egressVirtualServiceForExternalService(serviceName))
	result = append(result, meshVirtualServiceForExternalService(serviceName))
	result = append(result, egressExternServiceEntryForExternalService(serviceName))
	return result
}

func CreateEntriesForExternalServiceClient(serviceName string, hostName string, serviceIP string, port int, namespace string, systemDomain string) []istioModel.Config {
	var configs []istioModel.Config

	serviceEntry := createEgressExternServiceEntryForExternalService(hostName, uint32(port), serviceName, namespace)
	configs = append(configs, serviceEntry)

	virtualService := createMeshVirtualServiceForExternalService(hostName, 443, serviceName, serviceIP, namespace)
	configs = append(configs, virtualService)

	virtualService = createEgressVirtualServiceForExternalService(hostName, uint32(port), serviceName, 443, namespace)
	configs = append(configs, virtualService)

	gateway := createEgressGatewayForExternalService(hostName, 443, serviceName, namespace)
	configs = append(configs, gateway)

	destinationRule := createEgressDestinationRuleForExternalService(hostName, uint32(port), serviceName, namespace, systemDomain)
	configs = append(configs, destinationRule)

	destinationRule = createSidecarDestinationRuleForExternalService(hostName, serviceName, namespace)
	configs = append(configs, destinationRule)

	return configs
}

func ToYamlDocuments(entry []istioModel.Config) (string, error) {
	var result, text string
	var err error

	for _, element := range entry {
		text, err = enrichAndtoText(element)
		result += "---\n" + text
	}

	return result, err
}

func enrichAndtoText(config istioModel.Config) (string, error) {
	kubernetesConf, err := ToRuntimeObject(config)
	if err != nil {
		return "", err
	}
	bytes, err := yaml.Marshal(kubernetesConf)
	return string(bytes), err
}

func ToRuntimeObject(config istioModel.Config) (crd.IstioObject, error) {
	schema := schemas[config.Type]
	return crd.ConvertConfig(schema, config)
}
