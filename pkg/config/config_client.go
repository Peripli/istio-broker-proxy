package config

import (
	"fmt"
	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/config/kube/crd"
	"istio.io/istio/pilot/pkg/model"
)

func createEgressVirtualServiceForExternalService(hostName string, port uint32, serviceName string, gatewayPort uint32, namespace string) model.Config {
	gatewayName := egressVirtualServiceForExternalService(serviceName).Name
	gatewayHost := fmt.Sprintf("istio-egressgateway-%s", serviceName)
	matchGateways := []string{gatewayHost}
	match := v1alpha3.L4MatchAttributes{Gateways: matchGateways, Port: gatewayPort}
	config := createGeneralVirtualServiceForExternalService(hostName, port, serviceName, gatewayName, gatewayHost, match, hostName)

	return enrichWithIstioDefaults(config, namespace)
}

func egressVirtualServiceForExternalService(serviceName string) ServiceId {
	return ServiceId{model.VirtualService.Type, fmt.Sprintf("egress-gateway-%s", serviceName)}
}

func createMeshVirtualServiceForExternalService(hostName string, port uint32, serviceName string, serviceIP string, namespace string) model.Config {
	gatewayName := meshVirtualServiceForExternalService(serviceName).Name
	matchGateways := []string{"mesh"}
	match := v1alpha3.L4MatchAttributes{Gateways: matchGateways, DestinationSubnets: []string{serviceIP}}
	config := createGeneralVirtualServiceForExternalService(serviceName, port, serviceName, gatewayName, "mesh", match, "istio-egressgateway.istio-system.svc.cluster.local")

	return enrichWithIstioDefaults(config, namespace)
}

func meshVirtualServiceForExternalService(serviceName string) ServiceId {
	return ServiceId{model.VirtualService.Type, fmt.Sprintf("mesh-to-egress-%s", serviceName)}
}

func enrichWithIstioDefaults(config model.Config, namespace string) model.Config {
	schema := schemas[config.Type]
	config.Namespace = namespace
	istioObject, _ := crd.ConvertConfig(schema, config)
	enrichedConfig, _ := crd.ConvertObject(schema, istioObject, "")
	return *enrichedConfig
}

func createGeneralVirtualServiceForExternalService(hostName string, port uint32, serviceName string, gatewayName string, gatewayHost string, match v1alpha3.L4MatchAttributes, destinationHost string) model.Config {
	destination := v1alpha3.Destination{Host: destinationHost, Port: &v1alpha3.PortSelector{Port: &v1alpha3.PortSelector_Number{Number: port}}, Subset: serviceName}
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

func createEgressGatewayForExternalService(hostName string, portNumber uint32, serviceName string, namespace string) model.Config {
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
	gatewaySpec := v1alpha3.Gateway{Selector: selector, Servers: []*v1alpha3.Server{{Port: &port, Hosts: hosts, Tls: &tls}}}
	config := model.Config{Spec: &gatewaySpec, ConfigMeta: model.ConfigMeta{Labels: map[string]string{"service": serviceName}}}

	config.Type = gateway
	config.Name = egressGatewayForExternalService(serviceName).Name

	return enrichWithIstioDefaults(config, namespace)
}

func egressGatewayForExternalService(serviceName string) ServiceId {
	return ServiceId{model.Gateway.Type, fmt.Sprintf("istio-egressgateway-%s", serviceName)}
}

func createEgressDestinationRuleForExternalService(hostName string, portNumber uint32, serviceName string, namespace string) model.Config {
	port := v1alpha3.PortSelector{Port: &v1alpha3.PortSelector_Number{Number: portNumber}}
	tls := createTlsSettings(hostName)
	portLevelSettings := []*v1alpha3.TrafficPolicy_PortTrafficPolicy{{Tls: &tls, Port: &port}}
	trafficPolicy := v1alpha3.TrafficPolicy{PortLevelSettings: portLevelSettings}
	subsets := []*v1alpha3.Subset{{Name: serviceName, TrafficPolicy: &trafficPolicy}}
	destinationRuleSpec := v1alpha3.DestinationRule{Host: hostName, Subsets: subsets}
	config := model.Config{Spec: &destinationRuleSpec}
	config.Type = destinationRule
	config.Name = egressDestinationRuleForExternalService(serviceName).Name

	return enrichWithIstioDefaults(config, namespace)
}

func egressDestinationRuleForExternalService(serviceName string) ServiceId {
	return ServiceId{model.DestinationRule.Type, fmt.Sprintf("egressgateway-%s", serviceName)}
}

func createTlsSettings(hostName string) v1alpha3.TLSSettings {
	certPath := "/etc/istio/egressgateway-certs/"
	caCertificate := certPath + "ca.crt"
	clientCertificate := certPath + "client.crt"
	privateKey := certPath + "client.key"
	sni := hostName
	// TODO. subjectAltName should correspond to  systemdomain: istio.cf.<context.landscape.domain>
	subjectAltNames := []string{"cf-service.services.cf.dev01.aws.istio.sapcloud.io",
		"istio.cf.dev01.aws.istio.sapcloud.io"}
	mode := v1alpha3.TLSSettings_MUTUAL
	tls := v1alpha3.TLSSettings{CaCertificates: caCertificate, ClientCertificate: clientCertificate, PrivateKey: privateKey,
		Sni: sni, SubjectAltNames: subjectAltNames, Mode: mode}

	return tls
}

func createEgressExternServiceEntryForExternalService(hostName string, portNumber uint32, serviceName string, namespace string) model.Config {
	portName := fmt.Sprintf("%s-port", serviceName)
	name := egressExternServiceEntryForExternalService(serviceName).Name

	config := createGeneralServiceEntryForExternalService(name, hostName, portNumber, portName, "TLS")

	return enrichWithIstioDefaults(config, namespace)
}

func egressExternServiceEntryForExternalService(serviceName string) ServiceId {
	return ServiceId{model.ServiceEntry.Type, fmt.Sprintf("%s-service", serviceName)}
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

func createSidecarDestinationRuleForExternalService(hostName string, serviceName string, namespace string) model.Config {
	sni := hostName
	mode := v1alpha3.TLSSettings_ISTIO_MUTUAL

	tls := v1alpha3.TLSSettings{Sni: sni, Mode: mode}
	trafficPolicy := v1alpha3.TrafficPolicy{Tls: &tls}
	subsets := []*v1alpha3.Subset{{Name: serviceName, TrafficPolicy: &trafficPolicy}}
	destinationRuleSpec := v1alpha3.DestinationRule{Host: "istio-egressgateway.istio-system.svc.cluster.local", Subsets: subsets}
	config := model.Config{Spec: &destinationRuleSpec}
	config.Type = destinationRule
	config.Name = sidecarDestinationRuleForExternalService(serviceName).Name
	return enrichWithIstioDefaults(config, namespace)
}

func sidecarDestinationRuleForExternalService(serviceName string) ServiceId {
	return ServiceId{model.DestinationRule.Type, fmt.Sprintf("sidecar-to-egress-%s", serviceName)}
}
