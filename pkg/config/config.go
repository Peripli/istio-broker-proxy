package config

import (
	"github.com/ghodss/yaml"
	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/config/kube/crd"
	"istio.io/istio/pilot/pkg/model"
	"strings"
)

type generatedServiceEntry struct {
	gateway        model.Config
	virtualService model.Config
	serviceEntry   model.Config
}

func CreateEntriesForExternalService(serviceName string, ipServiceEntry string, portServiceEntry uint32, hostVirtualService string, portVirtualService uint32) (string, error) {

	var entry generatedServiceEntry

	entry.gateway = createIngressGatewayForExternalService(hostVirtualService, 9000, serviceName, "client.istio.sapcloud.io")
	entry.virtualService = createVirtualServiceForExternalService(hostVirtualService, portVirtualService, serviceName)
	entry.serviceEntry = createServiceEntryForExternalService(ipServiceEntry, portServiceEntry, serviceName)

	return toYamlArray(entry)
}

func createServiceEntryForExternalService(endpointAddress string, port uint32, serviceName string) model.Config {
	hosts := []string{serviceName + ".service-fabrik"}
	//FIXME: Should all service names end with -server???
	shortName := strings.TrimSuffix(serviceName, "-server")

	ports := v1alpha3.Port{Number: port, Name: shortName, Protocol: "TCP"}
	resolution := v1alpha3.ServiceEntry_STATIC
	endpoint := v1alpha3.ServiceEntry_Endpoint{Address: endpointAddress}
	serviceEntry := v1alpha3.ServiceEntry{Hosts: hosts, Ports: []*v1alpha3.Port{&ports}, Resolution: resolution,
		Endpoints: []*v1alpha3.ServiceEntry_Endpoint{&endpoint}}
	config := model.Config{Spec: &serviceEntry}
	config.Name = serviceName + "-service-entry"

	return config
}

func createVirtualServiceForExternalService(hostName string, port uint32, serviceName string) model.Config {
	destination := v1alpha3.Destination{Host: serviceName + ".service-fabrik",
		Port: &v1alpha3.PortSelector{Port: &v1alpha3.PortSelector_Number{Number: port}}}
	route := v1alpha3.TCPRoute{Route: []*v1alpha3.DestinationWeight{&v1alpha3.DestinationWeight{Destination: &destination}}}
	tcpRoutes := []*v1alpha3.TCPRoute{&route}
	hosts := []string{hostName}
	gateways := []string{serviceName + "-gateway"}
	virtualService := v1alpha3.VirtualService{Tcp: tcpRoutes, Hosts: hosts, Gateways: gateways}
	config := model.Config{Spec: &virtualService}
	config.Name = serviceName + "-virtual-service"

	return config
}

func createIngressGatewayForExternalService(hostName string, portNumber uint32, serviceName string, clientName string) model.Config {
	port := v1alpha3.Port{Number: portNumber, Name: "tls", Protocol: "TLS"}
	hosts := []string{hostName}
	tls := v1alpha3.Server_TLSOptions{Mode: v1alpha3.Server_TLSOptions_MUTUAL,
		ServerCertificate: "/var/vcap/jobs/envoy/config/certs/" + serviceName + ".crt",
		PrivateKey:        "/var/vcap/jobs/envoy/config/certs/" + serviceName + ".key",
		CaCertificates:    "/var/vcap/jobs/envoy/config/certs/ca.crt",
		SubjectAltNames:   []string{clientName}}
	gateway := v1alpha3.Gateway{Servers: []*v1alpha3.Server{&v1alpha3.Server{Port: &port, Hosts: hosts, Tls: &tls}}}
	config := model.Config{Spec: &gateway}
	config.Name = serviceName + "-gateway"

	return config
}

func toYamlArray(entry generatedServiceEntry) (string, error) {
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
