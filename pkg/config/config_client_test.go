package config

import (
	"fmt"
	. "github.com/onsi/gomega"
	"testing"
)

func TestClientGatewayFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	gatewayConfig := createEgressGatewayForExternalService("mypostgres.istio.my.arbitrary.domain.io", 443, "mypostgres", "catalog")
	g.Expect(gatewayConfig.Version).To(Equal("v1alpha3"))
	g.Expect(gatewayConfig.Group).To(Equal("networking.istio.io"))
	g.Expect(gatewayConfig.Type).To(Equal("gateway"))
	g.Expect(gatewayConfig.Name).To(Equal("istio-egressgateway-mypostgres.catalog.svc.cluster.local"))
	g.Expect(gatewayConfig.Spec.String()).To(Equal(`servers:<port:<number:443 protocol:"TLS" name:"tcp-port-443" > hosts:"mypostgres.istio.my.arbitrary.domain.io" tls:<mode:MUTUAL server_certificate:"/etc/certs/cert-chain.pem" private_key:"/etc/certs/key.pem" ca_certificates:"/etc/certs/root-cert.pem" subject_alt_names:"spiffe://cluster.local/ns/default/sa/default" > > selector:<key:"istio" value:"egressgateway" > `))
}

func TestClientVirtualServiceFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	virtualServiceSpec := createEgressVirtualServiceForExternalService("mypostgres.istio.my.arbitrary.domain.io",
		9000,
		"mypostgres", 443, "catalog")

	g.Expect(virtualServiceSpec.Version).To(Equal("v1alpha3"))
	g.Expect(virtualServiceSpec.Group).To(Equal("networking.istio.io"))
	g.Expect(virtualServiceSpec.Type).To(Equal("virtual-service"))
	g.Expect(virtualServiceSpec.Name).To(Equal("egress-gateway-mypostgres"))
	g.Expect(virtualServiceSpec.Spec.String()).To(Equal(`hosts:"mypostgres.istio.my.arbitrary.domain.io" gateways:"istio-egressgateway-mypostgres.catalog.my.arbitrary.domain.io" tcp:<match:<port:443 gateways:"istio-egressgateway-mypostgres.istio.my.arbitrary.domain.io" > route:<destination:<host:"mypostgres.istio.my.arbitrary.domain.io" subset:"mypostgres" port:<number:9000 > > > > `))
}

func TestClientMeshVirtualServiceFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	virtualServiceConfig := createMeshVirtualServiceForExternalService("mypostgres.istio.my.arbitrary.domain.io",
		443,
		"mypostgres", "100.66.152.30", "catalog")

	g.Expect(virtualServiceConfig.Version).To(Equal("v1alpha3"))
	g.Expect(virtualServiceConfig.Group).To(Equal("networking.istio.io"))
	g.Expect(virtualServiceConfig.Type).To(Equal("virtual-service"))
	g.Expect(virtualServiceConfig.Name).To(Equal("mesh-to-egress-mypostgres"))
	g.Expect(virtualServiceConfig.Spec.String()).To(Equal(`hosts:"mypostgres" gateways:"mesh.istio.my.arbitrary.domain.io" tcp:<match:<destination_subnets:"100.66.152.30" gateways:"mesh" > route:<destination:<host:"istio-egressgateway.istio-system.svc.cluster.local" subset:"mypostgres" port:<number:443 > > > > `))
}

func TestClientEgressDestinationRuleFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	destinationRuleConfig := createEgressDestinationRuleForExternalService("mypostgres.istio.my.arbitrary.domain.io",
		9000,
		"mypostgres", "catalog", "istio.my.arbitrary.domain.io")

	g.Expect(destinationRuleConfig.Version).To(Equal("v1alpha3"))
	g.Expect(destinationRuleConfig.Group).To(Equal("networking.istio.io"))
	g.Expect(destinationRuleConfig.Type).To(Equal("destination-rule"))
	g.Expect(destinationRuleConfig.Name).To(Equal("egressgateway-mypostgres"))
	fmt.Println(destinationRuleConfig.Spec.String())
	g.Expect(destinationRuleConfig.Spec.String()).To(Equal(`host:"mypostgres.istio.my.arbitrary.domain.io" subsets:<name:"mypostgres" traffic_policy:<port_level_settings:<port:<number:9000 > tls:<mode:MUTUAL client_certificate:"/etc/istio/egressgateway-certs/client.crt" private_key:"/etc/istio/egressgateway-certs/client.key" ca_certificates:"/etc/istio/egressgateway-certs/ca.crt" subject_alt_names:"istio.my.arbitrary.domain.io" sni:"mypostgres.istio.my.arbitrary.domain.io" > > > > `))
}

func TestClientExternServiceEntryFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	serviceEntryConfig := createEgressExternServiceEntryForExternalService("mypostgres.istio.my.arbitrary.domain.io",
		9000,
		"mypostgres", "catalog")

	g.Expect(serviceEntryConfig.Version).To(Equal("v1alpha3"))
	g.Expect(serviceEntryConfig.Group).To(Equal("networking.istio.io"))
	g.Expect(serviceEntryConfig.Type).To(Equal("service-entry"))
	g.Expect(serviceEntryConfig.Name).To(Equal("mypostgres-service"))
	g.Expect(serviceEntryConfig.Spec.String()).To(Equal(`hosts:"mypostgres.istio.my.arbitrary.domain.io" ports:<number:9000 protocol:"TLS" name:"mypostgres-port" > resolution:DNS `))
}

func TestClientSidecarDestinationRuleFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	destinationRuleConfig := createSidecarDestinationRuleForExternalService("mypostgres.istio.my.arbitrary.domain.io",
		"mypostgres", "catalog")

	g.Expect(destinationRuleConfig.Version).To(Equal("v1alpha3"))
	g.Expect(destinationRuleConfig.Group).To(Equal("networking.istio.io"))
	g.Expect(destinationRuleConfig.Type).To(Equal("destination-rule"))
	g.Expect(destinationRuleConfig.Name).To(Equal("sidecar-to-egress-mypostgres"))
	g.Expect(destinationRuleConfig.Spec.String()).To(Equal(`host:"istio-egressgateway.istio-system.svc.cluster.local" subsets:<name:"mypostgres" traffic_policy:<tls:<mode:ISTIO_MUTUAL sni:"mypostgres.istio.my.arbitrary.domain.io" > > > `))

}
