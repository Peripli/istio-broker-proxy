package config

import (
	"fmt"
	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/model"
)

func createEgressVirtualServiceForExternalService(hostName string, port uint32, serviceName string, gatewayPort uint32) model.Config {
	gatewayName := fmt.Sprintf("egress-gateway-%s", serviceName)
	gatewayHost := fmt.Sprintf("istio-egressgateway-%s", serviceName)
	config := createGeneralVirtualServiceForExternalService(hostName, port, serviceName, gatewayName, gatewayHost, gatewayPort, hostName)
	return config
}

func createMeshVirtualServiceForExternalService(hostName string, port uint32, serviceName string, gatewayPort uint32) model.Config {
	gatewayName := fmt.Sprintf("direct-through-egress-mesh-%s", serviceName)
	config := createGeneralVirtualServiceForExternalService(hostName, port, serviceName, gatewayName, "mesh", gatewayPort, "istio-egressgateway.istio-system.svc.cluster.local")
	return config
}

func createGeneralVirtualServiceForExternalService(hostName string, port uint32, serviceName string, gatewayName string, gatewayHost string, gatewayPort uint32, destinationHost string) model.Config {
	destination := v1alpha3.Destination{Host: destinationHost, Port: &v1alpha3.PortSelector{Port: &v1alpha3.PortSelector_Number{Number: port}}, Subset: serviceName}
	matchGateways := []string{fmt.Sprintf(gatewayHost)}
	match := v1alpha3.L4MatchAttributes{Gateways: matchGateways, Port: gatewayPort}
	route := v1alpha3.TCPRoute{Route: []*v1alpha3.DestinationWeight{{Destination: &destination}}, Match: []*v1alpha3.L4MatchAttributes{&match}}
	tcpRoutes := []*v1alpha3.TCPRoute{&route}
	hosts := []string{hostName}
	gateways := []string{fmt.Sprintf(gatewayHost)}
	virtualServiceSpec := v1alpha3.VirtualService{Tcp: tcpRoutes, Hosts: hosts, Gateways: gateways}
	config := model.Config{Spec: &virtualServiceSpec}
	config.Type = virtualService
	config.Name = gatewayName

	return config
}

func createEgressGatewayForExternalService(hostName string, portNumber uint32, serviceName string) model.Config {
	portName := fmt.Sprintf("tcp-port-%d", portNumber)
	port := v1alpha3.Port{Number: portNumber, Name: portName, Protocol: "TLS"}
	hosts := []string{hostName}
	certPath := "/etc/certs/"
	tls := v1alpha3.Server_TLSOptions{Mode: v1alpha3.Server_TLSOptions_MUTUAL,
		ServerCertificate: certPath + "cert-chain.pem",
		PrivateKey:        certPath + "key.pem",
		CaCertificates:    certPath + "root-cert.pem",
		SubjectAltNames:   []string{"spiffe://cluster.local/ns/default/sa/default"}}
	selector := make(map[string]string)
	selector["istio"] = "egressgateway"
	gatewaySpec := v1alpha3.Gateway{Selector: selector, Servers: []*v1alpha3.Server{&v1alpha3.Server{Port: &port, Hosts: hosts, Tls: &tls}}}
	config := model.Config{Spec: &gatewaySpec}
	config.Type = gateway
	config.Name = fmt.Sprintf("istio-egressgateway-%s", serviceName)

	return config
}

func createEgressDestinationRuleForExternalService(hostName string, portNumber uint32, serviceName string) model.Config {
	port := v1alpha3.PortSelector{Port: &v1alpha3.PortSelector_Number{Number: portNumber}}
	tls := createTlsSettings(hostName)
	loadBalancer := v1alpha3.LoadBalancerSettings{&v1alpha3.LoadBalancerSettings_Simple{}}
	portLevelSettings := []*v1alpha3.TrafficPolicy_PortTrafficPolicy{{Tls: &tls, Port: &port}}
	trafficPolicy := v1alpha3.TrafficPolicy{PortLevelSettings: portLevelSettings, LoadBalancer: &loadBalancer}
	subsets := []*v1alpha3.Subset{{Name: serviceName, TrafficPolicy: &trafficPolicy}}
	destinationRuleSpec := v1alpha3.DestinationRule{Host: hostName, Subsets: subsets}
	config := model.Config{Spec: &destinationRuleSpec}
	config.Type = destinationRule
	config.Name = fmt.Sprintf("egressgateway-%s", serviceName)

	return config
}

func createTlsSettings(hostName string) v1alpha3.TLSSettings {
	certPath := "/etc/istio/egressgateway-certs/"
	caCertificate := certPath + "ca.crt"
	clientCertificate := certPath + "client.crt"
	privateKey := certPath + "client.key"
	sni := hostName
	subjectAltNames := []string{"postgres.services.cf.dev01.aws.istio.sapcloud.io"}
	mode := v1alpha3.TLSSettings_MUTUAL
	tls := v1alpha3.TLSSettings{CaCertificates: caCertificate, ClientCertificate: clientCertificate, PrivateKey: privateKey,
		Sni: sni, SubjectAltNames: subjectAltNames, Mode: mode}

	return tls
}

func createEgressInternServiceEntryForExternalService(hostName string, portNumber uint32, serviceName string) model.Config {
	portName := fmt.Sprintf("tcp-port-%d", portNumber)
	name := fmt.Sprintf("internal-services-%s", serviceName)

	config := createGeneralServiceEntryForExternalService(name, hostName, portNumber, portName, "TCP")

	return config
}

func createEgressExternServiceEntryForExternalService(hostName string, portNumber uint32, serviceName string) model.Config {
	portName := fmt.Sprintf("%s-port", serviceName)
	name := fmt.Sprintf("%s-service", serviceName)

	config := createGeneralServiceEntryForExternalService(name, hostName, portNumber, portName, "TLS")

	return config
}

func createGeneralServiceEntryForExternalService(serviceEntryName string, hostName string, portNumber uint32, portName string, protocol string) model.Config {
	resolution := v1alpha3.ServiceEntry_DNS
	hosts := []string{hostName}
	ports := v1alpha3.Port{Number: portNumber, Name: portName, Protocol: protocol}
	serviceEntrySpec := v1alpha3.ServiceEntry{Hosts: hosts, Ports: []*v1alpha3.Port{&ports}, Resolution: resolution}
	config := model.Config{Spec: &serviceEntrySpec}
	config.Type = serviceEntry
	config.Name = serviceEntryName

	return config
}
