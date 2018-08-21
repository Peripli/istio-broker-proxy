package config

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/gomega"
	"istio.io/istio/pilot/pkg/config/kube/crd"
	"istio.io/istio/pilot/pkg/model"
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
      privateKey: /var/vcap/jobs/envoy/config/certs/postgres-server.key
      serverCertificate: /var/vcap/jobs/envoy/config/certs/postgres-server.crt
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

	gateway := createIngressGatewayForExternalService("postgres.services.cf.dev01.aws.istio.sapcloud.io",
		9000,
		"postgres-server",
		"client.istio.sapcloud.io")

	text, err := toText(model.Gateway, gateway)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(gateway_ingress_yml))
}

func TestServerVirtualServiceFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	virtualService := createIngressVirtualServiceForExternalService("postgres.services.cf.dev01.aws.istio.sapcloud.io",
		47637,
		"postgres-server")

	text, err := toText(model.VirtualService, virtualService)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(virtual_service_ingress_yml))
}

func TestServiceEntryFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	virtualService := createServiceEntryForExternalService("10.11.241.0", 47637, "postgres-server")

	text, err := toText(model.ServiceEntry, virtualService)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(service_entry_yml))
}

func toText(schema model.ProtoSchema, config model.Config) (string, error) {
	kubernetesConf, err := crd.ConvertConfig(schema, config)
	if err != nil {
		return "", err
	}
	bytes, err := yaml.Marshal(kubernetesConf)
	return string(bytes), err
}
