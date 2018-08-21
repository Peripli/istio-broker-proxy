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
	virtualService := v1alpha3.VirtualService{Tcp: tcpRoutes, Hosts: hosts, Gateways: gateways}
	config := model.Config{Spec: &virtualService}
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
	gateway := v1alpha3.Gateway{Selector: selector, Servers: []*v1alpha3.Server{&v1alpha3.Server{Port: &port, Hosts: hosts, Tls: &tls}}}
	config := model.Config{Spec: &gateway}
	config.Name = fmt.Sprintf("istio-egressgateway-%s", serviceName)

	return config
}

func createEgressDestinationRuleForExternalService(hostName string, portNumber uint32, serviceName string) model.Config {
	portSelector := v1alpha3.PortSelector_Number{portNumber}
	port := v1alpha3.PortSelector{}
	//TODO: How to do it easier
	bytes := make([]byte, 100)
	portSelector.MarshalTo(bytes)
	port.Unmarshal(bytes)
	certPath := "/etc/istio/egressgateway-certs/"
	caCertificate := certPath + "ca.crt"
	clientCertificate := certPath + "client.crt"
	privateKey := certPath + "client.key"
	sni := hostName
	subjectAltNames := []string{"postgres.services.cf.dev01.aws.istio.sapcloud.io"}
	mode := v1alpha3.TLSSettings_MUTUAL
	tls := v1alpha3.TLSSettings{CaCertificates: caCertificate, ClientCertificate: clientCertificate, PrivateKey: privateKey,
		Sni: sni, SubjectAltNames: subjectAltNames, Mode: mode}
	loadBalancerSimple := v1alpha3.LoadBalancerSettings_Simple{}
	bytes = make([]byte, 100)
	loadBalancerSimple.MarshalTo(bytes)
	loadBalancer := v1alpha3.LoadBalancerSettings{}
	loadBalancer.Unmarshal(bytes)
	portLevelSettings := []*v1alpha3.TrafficPolicy_PortTrafficPolicy{{Tls: &tls, Port: &port}}
	trafficPolicy := v1alpha3.TrafficPolicy{PortLevelSettings: portLevelSettings, LoadBalancer: &loadBalancer}
	subsets := []*v1alpha3.Subset{{Name: serviceName, TrafficPolicy: &trafficPolicy}}
	destinationRule := v1alpha3.DestinationRule{Host: hostName, Subsets: subsets}
	config := model.Config{Spec: &destinationRule}
	config.Name = fmt.Sprintf("egressgateway-%s", serviceName)

	return config
}
