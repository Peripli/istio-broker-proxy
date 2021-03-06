package config

import (
	. "github.com/onsi/gomega"
	"testing"
)

const (
	gatewayIngressYml = `apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  creationTimestamp: null
  name: postgres-server-gateway
  namespace: default
spec:
  servers:
  - hosts:
    - postgres.istio.my.arbitrary.domain.io
    port:
      name: tls
      number: 9000
      protocol: TLS
    tls:
      caCertificates: /etc/istio/certs/ca.crt
      mode: MUTUAL
      privateKey: /etc/istio/certs/cf-service.key
      serverCertificate: /etc/istio/certs/cf-service.crt
      subjectAltNames:
      - client.my.client.domain.io
`
	virtualServiceIngressYml = `apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  creationTimestamp: null
  name: postgres-server-virtual-service
  namespace: default
spec:
  gateways:
  - postgres-server-gateway
  hosts:
  - postgres.istio.my.arbitrary.domain.io
  tcp:
  - route:
    - destination:
        host: postgres-server.service-fabrik
        port:
          number: 47637
`
	staticServiceEntryYml = `apiVersion: networking.istio.io/v1alpha3
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
	dnsServiceEntryYml = `apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  creationTimestamp: null
  name: postgres-server-service-entry
  namespace: default
spec:
  endpoints:
  - address: postgres.example.com
  hosts:
  - postgres-server.service-fabrik
  ports:
  - name: postgres-server-47637
    number: 47637
    protocol: TCP
  resolution: DNS
`
)

func TestServerGatewayFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	gatewaySpec := createIngressGatewayForExternalService("postgres.istio.my.arbitrary.domain.io",
		9000,
		"postgres-server",
		"client.my.client.domain.io")

	text, err := enrichAndtoText(gatewaySpec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(gatewayIngressYml))
}

func TestServerVirtualServiceFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	virtualServiceSpec := createIngressVirtualServiceForExternalService("postgres.istio.my.arbitrary.domain.io",
		47637,
		"postgres-server")

	text, err := enrichAndtoText(virtualServiceSpec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(virtualServiceIngressYml))
}

func TestStaticServiceEntryFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	serviceEntrySpec := createServiceEntryForExternalService("10.11.241.0", 47637, "postgres-server")

	text, err := enrichAndtoText(serviceEntrySpec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(staticServiceEntryYml))
	g.Expect(serviceEntrySpec.Type).To(Equal(serviceEntry))
}

func TestDNSServiceEntryFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	serviceEntrySpec := createServiceEntryForExternalService("postgres.example.com", 47637, "postgres-server")

	text, err := enrichAndtoText(serviceEntrySpec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(dnsServiceEntryYml))
	g.Expect(serviceEntrySpec.Type).To(Equal(serviceEntry))
}

