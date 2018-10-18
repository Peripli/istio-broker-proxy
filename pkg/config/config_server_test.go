package config

import (
	. "github.com/onsi/gomega"
	"testing"
)

const (
	gateway_ingress_yml = `apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  creationTimestamp: null
  name: postgres-server-gateway
  namespace: default
spec:
  servers:
  - hosts:
    - postgres.services.cf.dev01.aws.istio.sapcloud.io
    port:
      name: tls
      number: 9000
      protocol: TLS
    tls:
      caCertificates: /var/vcap/jobs/envoy/config/certs/ca.crt
      mode: MUTUAL
      privateKey: /var/vcap/jobs/envoy/config/certs/cf-service.key
      serverCertificate: /var/vcap/jobs/envoy/config/certs/cf-service.crt
      subjectAltNames:
      - client.istio.sapcloud.io
`
	virtual_service_ingress_yml = `apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  creationTimestamp: null
  name: postgres-server-virtual-service
  namespace: default
spec:
  gateways:
  - postgres-server-gateway
  hosts:
  - postgres.services.cf.dev01.aws.istio.sapcloud.io
  tcp:
  - route:
    - destination:
        host: postgres-server.service-fabrik
        port:
          number: 47637
`
	service_entry_yml = `apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  creationTimestamp: null
  name: postgres-server-service-entry
  namespace: default
spec:
  endpoints:
  - address: 10.11.241.0
  hosts:
  - postgres-server.service-fabrik
  ports:
  - name: postgres-server-47637
    number: 47637
    protocol: TCP
  resolution: STATIC
`
)

func TestServerGatewayFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	gatewaySpec := createIngressGatewayForExternalService("postgres.services.cf.dev01.aws.istio.sapcloud.io",
		9000,
		"postgres-server",
		"client.istio.sapcloud.io")

	text, err := enrichAndtoText(gatewaySpec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(gateway_ingress_yml))
}

func TestServerVirtualServiceFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	virtualServiceSpec := createIngressVirtualServiceForExternalService("postgres.services.cf.dev01.aws.istio.sapcloud.io",
		47637,
		"postgres-server")

	text, err := enrichAndtoText(virtualServiceSpec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(virtual_service_ingress_yml))
}

func TestServiceEntryFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	serviceEntrySpec := createServiceEntryForExternalService("10.11.241.0", 47637, "postgres-server")

	text, err := enrichAndtoText(serviceEntrySpec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(service_entry_yml))
	g.Expect(serviceEntrySpec.Type).To(Equal(serviceEntry))
}
