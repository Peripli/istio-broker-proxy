package config

import (
	"fmt"
	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/model"
)

const ingressCertName = "cf-service"

func createServiceEntryForExternalService(endpointAddress string, portNumber uint32, serviceName string) model.Config {
	hosts := []string{createServiceHost(serviceName)}
	portName := fmt.Sprintf("%s-%d", serviceName, portNumber)

	ports := v1alpha3.Port{Number: portNumber, Name: portName, Protocol: "TCP"}
	resolution := v1alpha3.ServiceEntry_STATIC
	endpoint := v1alpha3.ServiceEntry_Endpoint{Address: endpointAddress}
	serviceEntrySpec := v1alpha3.ServiceEntry{Hosts: hosts, Ports: []*v1alpha3.Port{&ports}, Resolution: resolution,
		Endpoints: []*v1alpha3.ServiceEntry_Endpoint{&endpoint}}
	config := model.Config{Spec: &serviceEntrySpec}
	config.Type = serviceEntry
	config.Name = serviceName + "-service-entry"

	return config
}

func createServiceHost(serviceName string) string {
	serviceHost := serviceName + ".service-fabrik"
	return serviceHost
}

func createIngressVirtualServiceForExternalService(hostName string, port uint32, serviceName string) model.Config {
	destination := v1alpha3.Destination{Host: createServiceHost(serviceName),
		Port: &v1alpha3.PortSelector{Port: &v1alpha3.PortSelector_Number{Number: port}}}
	route := v1alpha3.TCPRoute{Route: []*v1alpha3.DestinationWeight{&v1alpha3.DestinationWeight{Destination: &destination}}}
	tcpRoutes := []*v1alpha3.TCPRoute{&route}
	hosts := []string{hostName}
	gateways := []string{serviceName + "-gateway"}
	virtualServiceSpec := v1alpha3.VirtualService{Tcp: tcpRoutes, Hosts: hosts, Gateways: gateways}
	config := model.Config{Spec: &virtualServiceSpec}
	config.Type = virtualService
	config.Name = serviceName + "-virtual-service"

	return config
}

func createIngressGatewayForExternalService(hostName string, portNumber uint32, serviceName string, clientName string) model.Config {
	port := v1alpha3.Port{Number: portNumber, Name: "tls", Protocol: "TLS"}
	hosts := []string{hostName}
	certPath := "/var/vcap/jobs/envoy/config/certs/"
	tls := v1alpha3.Server_TLSOptions{Mode: v1alpha3.Server_TLSOptions_MUTUAL,
		ServerCertificate: certPath + ingressCertName + ".crt",
		PrivateKey:        certPath + ingressCertName + ".key",
		CaCertificates:    certPath + "ca.crt",
		SubjectAltNames:   []string{clientName}}
	gatewaySpec := v1alpha3.Gateway{Servers: []*v1alpha3.Server{&v1alpha3.Server{Port: &port, Hosts: hosts, Tls: &tls}}}
	config := model.Config{Spec: &gatewaySpec}
	config.Type = gateway
	config.Name = serviceName + "-gateway"

	return config
}
