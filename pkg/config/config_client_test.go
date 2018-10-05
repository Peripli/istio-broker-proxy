package config

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestClientGatewayFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	gatewayConfig := createEgressGatewayForExternalService("mypostgres.services.cf.dev01.aws.istio.sapcloud.io", 443, "mypostgres")
	g.Expect(gatewayConfig.Version).To(Equal("v1alpha3"))
	g.Expect(gatewayConfig.Group).To(Equal("networking.istio.io"))
	g.Expect(gatewayConfig.Type).To(Equal("gateway"))
	g.Expect(gatewayConfig.Name).To(Equal("istio-egressgateway-mypostgres"))
	g.Expect(gatewayConfig.Spec.String()).To(Equal(`servers:<port:<number:443 protocol:"TLS" name:"tcp-port-443" > hosts:"mypostgres.services.cf.dev01.aws.istio.sapcloud.io" tls:<mode:MUTUAL server_certificate:"/etc/certs/cert-chain.pem" private_key:"/etc/certs/key.pem" ca_certificates:"/etc/certs/root-cert.pem" subject_alt_names:"spiffe://cluster.local/ns/default/sa/default" > > selector:<key:"istio" value:"egressgateway" > `))
}

func TestClientVirtualServiceFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	virtualServiceSpec := createEgressVirtualServiceForExternalService("mypostgres.services.cf.dev01.aws.istio.sapcloud.io",
		9000,
		"mypostgres", 443)

	g.Expect(virtualServiceSpec.Version).To(Equal("v1alpha3"))
	g.Expect(virtualServiceSpec.Group).To(Equal("networking.istio.io"))
	g.Expect(virtualServiceSpec.Type).To(Equal("virtual-service"))
	g.Expect(virtualServiceSpec.Name).To(Equal("egress-gateway-mypostgres"))
	g.Expect(virtualServiceSpec.Spec.String()).To(Equal(`hosts:"mypostgres" gateways:"istio-egressgateway-mypostgres" tcp:<match:<port:443 gateways:"istio-egressgateway-mypostgres" > route:<destination:<host:"mypostgres.services.cf.dev01.aws.istio.sapcloud.io" subset:"mypostgres" port:<number:9000 > > > > `))
}

func TestClientMeshVirtualServiceFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	virtualServiceConfig := createMeshVirtualServiceForExternalService("mypostgres.services.cf.dev01.aws.istio.sapcloud.io",
		443,
		"mypostgres", "100.66.152.30")

	g.Expect(virtualServiceConfig.Version).To(Equal("v1alpha3"))
	g.Expect(virtualServiceConfig.Group).To(Equal("networking.istio.io"))
	g.Expect(virtualServiceConfig.Type).To(Equal("virtual-service"))
	g.Expect(virtualServiceConfig.Name).To(Equal("direct-through-egress-mesh-mypostgres"))
	g.Expect(virtualServiceConfig.Spec.String()).To(Equal(`hosts:"mypostgres" gateways:"mesh" tcp:<match:<destination_subnets:"100.66.152.30" gateways:"mesh" > route:<destination:<host:"istio-egressgateway.istio-system.svc.cluster.local" subset:"mypostgres" port:<number:443 > > > > `))
}

func TestClientEgressDestinationRuleFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	destinationRuleConfig := createEgressDestinationRuleForExternalService("mypostgres.services.cf.dev01.aws.istio.sapcloud.io",
		9000,
		"mypostgres")

	g.Expect(destinationRuleConfig.Version).To(Equal("v1alpha3"))
	g.Expect(destinationRuleConfig.Group).To(Equal("networking.istio.io"))
	g.Expect(destinationRuleConfig.Type).To(Equal("destination-rule"))
	g.Expect(destinationRuleConfig.Name).To(Equal("egressgateway-mypostgres"))
	g.Expect(destinationRuleConfig.Spec.String()).To(Equal(`host:"mypostgres.services.cf.dev01.aws.istio.sapcloud.io" subsets:<name:"mypostgres" traffic_policy:<port_level_settings:<port:<number:9000 > tls:<mode:MUTUAL client_certificate:"/etc/istio/egressgateway-certs/client.crt" private_key:"/etc/istio/egressgateway-certs/client.key" ca_certificates:"/etc/istio/egressgateway-certs/ca.crt" subject_alt_names:"mypostgres.services.cf.dev01.aws.istio.sapcloud.io" sni:"mypostgres.services.cf.dev01.aws.istio.sapcloud.io" > > > > `))
}

func TestClientExternServiceEntryFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	serviceEntryConfig := createEgressExternServiceEntryForExternalService("mypostgres.services.cf.dev01.aws.istio.sapcloud.io",
		9000,
		"mypostgres")

	g.Expect(serviceEntryConfig.Version).To(Equal("v1alpha3"))
	g.Expect(serviceEntryConfig.Group).To(Equal("networking.istio.io"))
	g.Expect(serviceEntryConfig.Type).To(Equal("service-entry"))
	g.Expect(serviceEntryConfig.Name).To(Equal("mypostgres-service"))
	g.Expect(serviceEntryConfig.Spec.String()).To(Equal(`hosts:"mypostgres.services.cf.dev01.aws.istio.sapcloud.io" ports:<number:9000 protocol:"TLS" name:"mypostgres-port" > resolution:DNS `))
}

func TestClientSidecarDestinationRuleFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	destinationRuleConfig := createSidecarDestinationRuleForExternalService("mypostgres.services.cf.dev01.aws.istio.sapcloud.io",
		"mypostgres")

	g.Expect(destinationRuleConfig.Version).To(Equal("v1alpha3"))
	g.Expect(destinationRuleConfig.Group).To(Equal("networking.istio.io"))
	g.Expect(destinationRuleConfig.Type).To(Equal("destination-rule"))
	g.Expect(destinationRuleConfig.Name).To(Equal("sidecar-to-egress-mypostgres"))
	g.Expect(destinationRuleConfig.Spec.String()).To(Equal(`host:"istio-egressgateway.istio-system.svc.cluster.local" subsets:<name:"mypostgres" traffic_policy:<tls:<mode:ISTIO_MUTUAL sni:"mypostgres.services.cf.dev01.aws.istio.sapcloud.io" > > > `))

}
