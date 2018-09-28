package config

import (
	. "github.com/onsi/gomega"
	"testing"
)

const (
	service_entry_extern_yml = `apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  creationTimestamp: null
  name: mypostgres-service
  namespace: default
spec:
  hosts:
  - mypostgres.services.cf.dev01.aws.istio.sapcloud.io
  ports:
  - name: mypostgres-port
    number: 9000
    protocol: TLS
  resolution: DNS
`

	destination_rule_sidecar_yml = `apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  creationTimestamp: null
  name: sidecar-to-egress-mypostgres
  namespace: default
spec:
  host: istio-egressgateway.istio-system.svc.cluster.local
  subsets:
  - name: mypostgres
    trafficPolicy:
      tls:
        mode: ISTIO_MUTUAL
        sni: mypostgres.services.cf.dev01.aws.istio.sapcloud.io
`

	destination_rule_egress_yml = `apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  creationTimestamp: null
  name: egressgateway-mypostgres
  namespace: default
spec:
  host: mypostgres.services.cf.dev01.aws.istio.sapcloud.io
  subsets:
  - name: mypostgres
    trafficPolicy:
      portLevelSettings:
      - port:
          number: 9000
        tls:
          caCertificates: /etc/istio/egressgateway-certs/ca.crt
          clientCertificate: /etc/istio/egressgateway-certs/client.crt
          mode: MUTUAL
          privateKey: /etc/istio/egressgateway-certs/client.key
          sni: mypostgres.services.cf.dev01.aws.istio.sapcloud.io
          subjectAltNames:
          - mypostgres.services.cf.dev01.aws.istio.sapcloud.io
`

	virtual_service_mesh_yml = `apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  creationTimestamp: null
  name: direct-through-egress-mesh-mypostgres
  namespace: default
spec:
  gateways:
  - mesh
  hosts:
  - mypostgres
  tcp:
  - match:
    - destinationSubnets:
      - 100.66.152.30
      gateways:
      - mesh
    route:
    - destination:
        host: istio-egressgateway.istio-system.svc.cluster.local
        port:
          number: 443
        subset: mypostgres
`
	virtual_service_egress_yml = `apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  creationTimestamp: null
  name: egress-gateway-mypostgres
  namespace: default
spec:
  gateways:
  - istio-egressgateway-mypostgres
  hosts:
  - mypostgres
  tcp:
  - match:
    - gateways:
      - istio-egressgateway-mypostgres
      port: 443
    route:
    - destination:
        host: mypostgres.services.cf.dev01.aws.istio.sapcloud.io
        port:
          number: 9000
        subset: mypostgres
`
	gateway_egress_yml = `apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  creationTimestamp: null
  name: istio-egressgateway-mypostgres
  namespace: default
spec:
  selector:
    istio: egressgateway
  servers:
  - hosts:
    - mypostgres.services.cf.dev01.aws.istio.sapcloud.io
    port:
      name: tcp-port-443
      number: 443
      protocol: TLS
    tls:
      caCertificates: /etc/certs/root-cert.pem
      mode: MUTUAL
      privateKey: /etc/certs/key.pem
      serverCertificate: /etc/certs/cert-chain.pem
      subjectAltNames:
      - spiffe://cluster.local/ns/default/sa/default
`
)

func TestClientGatewayFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	gatewaySpec := createEgressGatewayForExternalService("mypostgres.services.cf.dev01.aws.istio.sapcloud.io", 443, "mypostgres")

	text, err := toText(gatewaySpec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(gateway_egress_yml))
}

func TestClientVirtualServiceFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	virtualServiceSpec := createEgressVirtualServiceForExternalService("mypostgres.services.cf.dev01.aws.istio.sapcloud.io",
		9000,
		"mypostgres", 443)

	text, err := toText(virtualServiceSpec)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(virtual_service_egress_yml))
}

func TestClientMeshVirtualServiceFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	virtualServiceSpec := createMeshVirtualServiceForExternalService("mypostgres.services.cf.dev01.aws.istio.sapcloud.io",
		443,
		"mypostgres", "100.66.152.30")

	text, err := toText(virtualServiceSpec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(virtual_service_mesh_yml))
}

func TestClientEgressDestinationRuleFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	destinationRuleSpec := createEgressDestinationRuleForExternalService("mypostgres.services.cf.dev01.aws.istio.sapcloud.io",
		9000,
		"mypostgres")

	text, err := toText(destinationRuleSpec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(destination_rule_egress_yml))
}

func TestClientExternServiceEntryFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	serviceEntrySpec := createEgressExternServiceEntryForExternalService("mypostgres.services.cf.dev01.aws.istio.sapcloud.io",
		9000,
		"mypostgres")

	text, err := toText(serviceEntrySpec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(service_entry_extern_yml))
}

func TestClientSidecarDestinationRuleFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	destinationRuleSpec := createSidecarDestinationRuleForExternalService("mypostgres.services.cf.dev01.aws.istio.sapcloud.io",
		"mypostgres")

	text, err := toText(destinationRuleSpec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(destination_rule_sidecar_yml))
}
