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
	istioGateway         = "gateway"
	istioServiceEntry    = "service-entry"
	istioVirtualService  = "virtual-service"
	istioDestinationRule = "destination-rule"
)

var invalidIdentifiers = regexp.MustCompile(`[^0-9a-z-]`)

var schemas = map[string]istioModel.ProtoSchema{
	gateway:              istioModel.Gateway,
	serviceEntry:         istioModel.ServiceEntry,
	virtualService:       istioModel.VirtualService,
	destinationRule:      istioModel.DestinationRule,
	istioGateway:         istioModel.Gateway,
	istioServiceEntry:    istioModel.ServiceEntry,
	istioVirtualService:  istioModel.VirtualService,
	istioDestinationRule: istioModel.DestinationRule,
}

//IstioObjectID identifies an istio configuration object (e.g. a gateway with name XY)
type IstioObjectID struct {
	Type string
	Name string
}

//CreateEntriesForExternalService creates routing rules (gateway, virtual service, service entry) to route to a service with serviceName (provider side)
func CreateEntriesForExternalService(serviceName string, endpointServiceEntry string, portServiceEntry uint32, hostVirtualService string, clientName string, ingressPort uint32, providerSAN string) []istioModel.Config {
	var configs []istioModel.Config

	configs = append(configs, createIngressGatewayForExternalService(hostVirtualService, ingressPort, serviceName, clientName))
	configs = append(configs, createIngressVirtualServiceForExternalService(hostVirtualService, portServiceEntry, serviceName))
	configs = append(configs, createServiceEntryForExternalService(endpointServiceEntry, portServiceEntry, serviceName))

	return configs
}

//CreateIstioConfigForProvider creates istio routing rules for provider
func CreateIstioConfigForProvider(request *model.BindRequest, response *model.BindResponse, bindingID string, systemDomain string, providerSAN string) []istioModel.Config {
	var istioConfig []istioModel.Config
	consumerID := request.NetworkData.Data.ConsumerID
	for index, endpoint := range response.Endpoints {
		portServiceEntry := uint32(endpoint.Port)
		ingressPort := uint32(9000)

		serviceName := createValidIdentifer(fmt.Sprintf("%d-%s", index, bindingID))
		endpointServiceEntry := endpoint.Host
		hostVirtualService := profiles.CreateEndpointHosts(bindingID, systemDomain, index)
		istioConfig = append(istioConfig,
			CreateEntriesForExternalService(serviceName, endpointServiceEntry, portServiceEntry, hostVirtualService, consumerID, ingressPort, providerSAN)...)
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

//CreateEntriesForExternalServiceClient creates istio routing config for a service for the consumer side
func CreateEntriesForExternalServiceClient(serviceName string, hostName string, serviceIP string, port int, systemDomain string) []istioModel.Config {
	var configs []istioModel.Config

	serviceEntry := createEgressExternServiceEntryForExternalService(hostName, uint32(port), serviceName)
	configs = append(configs, serviceEntry)

	virtualService := createMeshVirtualServiceForExternalService(hostName, 443, serviceName, serviceIP)
	configs = append(configs, virtualService)

	virtualService = createEgressVirtualServiceForExternalService(hostName, uint32(port), serviceName, 443)
	configs = append(configs, virtualService)

	gateway := createEgressGatewayForExternalService(hostName, 443, serviceName)
	configs = append(configs, gateway)

	destinationRule := createEgressDestinationRuleForExternalService(hostName, uint32(port), serviceName, systemDomain)
	configs = append(configs, destinationRule)

	destinationRule = createSidecarDestinationRuleForExternalService(hostName, serviceName)
	configs = append(configs, destinationRule)

	return configs
}

//ToYamlDocuments creates yaml config files
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
	kubernetesConf, err := toRuntimeObject(config)
	if err != nil {
		return "", err
	}
	bytes, err := yaml.Marshal(kubernetesConf)
	return string(bytes), err
}

func toRuntimeObject(config istioModel.Config) (crd.IstioObject, error) {
	schema := schemas[config.Type]
	return crd.ConvertConfig(schema, config)
}
