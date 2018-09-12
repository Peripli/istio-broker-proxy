package config

import (
	"fmt"
	"github.com/ghodss/yaml"
	"istio.io/istio/pilot/pkg/config/kube/crd"
	"istio.io/istio/pilot/pkg/model"
	"os"
	"path"
)

const (
	gateway         = "Gateway"
	serviceEntry    = "ServiceEntry"
	virtualService  = "VirtualService"
	destinationRule = "DestinationRule"
)

var schemas = map[string]model.ProtoSchema{
	gateway:         model.Gateway,
	serviceEntry:    model.ServiceEntry,
	virtualService:  model.VirtualService,
	destinationRule: model.DestinationRule,
}

func CreateEntriesForExternalService(serviceName string, endpointServiceEntry string, portServiceEntry uint32, hostVirtualService string, clientName string, ingressPort uint32) []model.Config {
	var configs []model.Config

	configs = append(configs, createIngressGatewayForExternalService(hostVirtualService, ingressPort, serviceName, clientName))
	configs = append(configs, createIngressVirtualServiceForExternalService(hostVirtualService, portServiceEntry, serviceName))
	configs = append(configs, createServiceEntryForExternalService(endpointServiceEntry, portServiceEntry, serviceName))

	return configs
}

func WriteIstioFilesForProvider(istioDirectory string, bindingId string) func([]byte, []byte) error {
	return func(request []byte, response []byte) error {
		file, err := os.Create(path.Join(istioDirectory, bindingId) + ".yml")
		if nil != err {
			return err
		}
		defer file.Close()

		//Todo: get values for parameters from request and response
		serviceInstanceId := ""
		originalEndpointHost := "" //responseBody cave! can be more than one!
		portServiceEntry := uint32(0)
		ingressDomain := ""

		consumerId := "147" //requestBody
		ingressPort := uint32(9000)

		serviceName := fmt.Sprintf("%s%v", serviceInstanceId, originalEndpointHost) //ToDo: orignalEndpointHost might be hashed
		endpointServiceEntry := originalEndpointHost
		hostVirtualService := fmt.Sprintf("%s%v%s", serviceInstanceId, originalEndpointHost, ingressDomain)

		fileContent, err := ToYamlDocuments(CreateEntriesForExternalService(serviceName, endpointServiceEntry, portServiceEntry, hostVirtualService, consumerId, ingressPort))
		if nil != err {
			return err
		}
		file.Write([]byte(fileContent))
		return nil
	}
}
func CreateEntriesForExternalServiceClient(serviceName string, hostName string, portNumber uint32) []model.Config {
	var configs []model.Config

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

func ToYamlDocuments(entry []model.Config) (string, error) {
	var result, text string
	var err error

	for _, element := range entry {
		text, err = toText(element)
		result += "---\n" + text
	}

	return result, err
}

func toText(config model.Config) (string, error) {
	schema := schemas[config.Type]
	kubernetesConf, err := crd.ConvertConfig(schema, config)
	if err != nil {
		return "", err
	}
	bytes, err := yaml.Marshal(kubernetesConf)
	return string(bytes), err
}
