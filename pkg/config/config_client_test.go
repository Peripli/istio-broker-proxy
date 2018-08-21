package config

import (
	. "github.com/onsi/gomega"
	"istio.io/istio/pilot/pkg/model"
	"testing"
)

const (
	service_entry_extern_yml = `apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  creationTimestamp: null
  name: postgres-service
  namespace: default
spec:
  hosts:
  - postgres.services.cf.dev01.aws.istio.sapcloud.io
  ports:
  - name: postgres-port
    number: 9000
    protocol: TLS
  resolution: DNS
`
	service_entry_intern_yml = `apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  creationTimestamp: null
  name: internal-services-postgres
  namespace: default
spec:
  hosts:
  - postgres.services.cf.dev01.aws.istio.sapcloud.io
  ports:
  - name: tcp-port-5556
    number: 5556
    protocol: TCP
  resolution: DNS
`
	destination_rule_egress_yml = `apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  creationTimestamp: null
  name: egressgateway-postgres
  namespace: default
spec:
  host: postgres.services.cf.dev01.aws.istio.sapcloud.io
  subsets:
  - name: postgres
    trafficPolicy:
      loadBalancer:
        simple: ROUND_ROBIN
      portLevelSettings:
      - port:
          number: 9000
        tls:
          caCertificates: /etc/istio/egressgateway-certs/ca.crt
          clientCertificate: /etc/istio/egressgateway-certs/client.crt
          mode: MUTUAL
          privateKey: /etc/istio/egressgateway-certs/client.key
          sni: postgres.services.cf.dev01.aws.istio.sapcloud.io
          subjectAltNames:
          - postgres.services.cf.dev01.aws.istio.sapcloud.io
`

	virtual_service_mesh_yml = `apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  creationTimestamp: null
  name: direct-through-egress-mesh-postgres
  namespace: default
spec:
  gateways:
  - mesh
  hosts:
  - postgres.services.cf.dev01.aws.istio.sapcloud.io
  tcp:
  - match:
    - gateways:
      - mesh
      port: 5556
    route:
    - destination:
        host: istio-egressgateway.istio-system.svc.cluster.local
        port:
          number: 443
        subset: postgres
`
	virtual_service_egress_yml = `apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  creationTimestamp: null
  name: egress-gateway-postgres
  namespace: default
spec:
  gateways:
  - istio-egressgateway-postgres
  hosts:
  - postgres.services.cf.dev01.aws.istio.sapcloud.io
  tcp:
  - match:
    - gateways:
      - istio-egressgateway-postgres
      port: 443
    route:
    - destination:
        host: postgres.services.cf.dev01.aws.istio.sapcloud.io
        port:
          number: 9000
        subset: postgres
`
	gateway_egress_yml = `apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  creationTimestamp: null
  name: istio-egressgateway-postgres
  namespace: default
spec:
  selector:
    istio: egressgateway
  servers:
  - hosts:
    - postgres.services.cf.dev01.aws.istio.sapcloud.io
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

	gateway := createEgressGatewayForExternalService("postgres.services.cf.dev01.aws.istio.sapcloud.io", 443, "postgres")

	text, err := toText(model.Gateway, gateway)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(gateway_egress_yml))
}

func TestClientVirtualServiceFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	virtualService := createEgressVirtualServiceForExternalService("postgres.services.cf.dev01.aws.istio.sapcloud.io",
		9000,
		"postgres", 443)

	text, err := toText(model.VirtualService, virtualService)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(virtual_service_egress_yml))
}

func TestClientMeshVirtualServiceFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	virtualService := createMeshVirtualServiceForExternalService("postgres.services.cf.dev01.aws.istio.sapcloud.io",
		443,
		"postgres", 5556)

	text, err := toText(model.VirtualService, virtualService)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(virtual_service_mesh_yml))
}

func TestClientEgressDestinationRuleFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	destinationRule := createEgressDestinationRuleForExternalService("postgres.services.cf.dev01.aws.istio.sapcloud.io",
		9000,
		"postgres")

	text, err := toText(model.DestinationRule, destinationRule)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(destination_rule_egress_yml))
}

func TestClientInternServiceEntryFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	serviceEntry := createEgressInternServiceEntryForExternalService("postgres.services.cf.dev01.aws.istio.sapcloud.io",
		5556,
		"postgres")

	text, err := toText(model.ServiceEntry, serviceEntry)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(service_entry_intern_yml))
}

func TestClientExternServiceEntryFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	serviceEntry := createEgressExternServiceEntryForExternalService("postgres.services.cf.dev01.aws.istio.sapcloud.io",
		9000,
		"postgres")

	text, err := toText(model.ServiceEntry, serviceEntry)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(service_entry_extern_yml))
}
