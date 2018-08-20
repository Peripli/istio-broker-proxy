package config

import (
	"errors"
	"github.com/ghodss/yaml"
	. "github.com/onsi/gomega"
	"istio.io/istio/pilot/pkg/config/kube/crd"
	"istio.io/istio/pilot/pkg/model"
	"testing"
)

const (
	gateway_yml = `apiVersion: networking.istio.io/v1alpha3
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
	virtual_service_yml = `apiVersion: networking.istio.io/v1alpha3
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
  - name: postgres
    number: 47637
    protocol: TCP
  resolution: STATIC
`
)

func TestCompleteEntryNotEmpty(t *testing.T) {
	g := NewGomegaWithT(t)

	configAsString, _ := CreateEntriesForExternalService("myservice", "10.10.10.10", 10, "myservice.landscape")

	var configObjects []interface{}
	yaml.Unmarshal([]byte(configAsString), &configObjects)

	g.Expect(configObjects).To(HaveLen(3))
}

func TestCompleteEntryGateway(t *testing.T) {
	g := NewGomegaWithT(t)

	configAsString, _ := CreateEntriesForExternalService("myservice", "10.10.10.10", 10, "myservice.landscape")

	var configObjects []interface{}
	yaml.Unmarshal([]byte(configAsString), &configObjects)

	gatewayConfig, err := lookupObjectFromConfig(configObjects, "Gateway")
	g.Expect(err).ShouldNot(HaveOccurred())

	gatewaySpec, _ := yaml.Marshal(gatewayConfig["spec"])
	gatewayMetadata, _ := yaml.Marshal(gatewayConfig["metadata"])

	g.Expect(string(gatewaySpec)).To(ContainSubstring("myservice.landscape"))
	g.Expect(string(gatewaySpec)).To(ContainSubstring("9000"))
	g.Expect(string(gatewaySpec)).To(ContainSubstring("client.istio.sapcloud.io"))
	g.Expect(string(gatewaySpec)).To(ContainSubstring("config/certs/myservice.key"))
	g.Expect(string(gatewaySpec)).To(ContainSubstring("config/certs/myservice.crt"))

	g.Expect(string(gatewayMetadata)).To(ContainSubstring("name: myservice-gateway"))
}

func TestCompleteServiceEntry(t *testing.T) {
	g := NewGomegaWithT(t)

	configAsString, _ := CreateEntriesForExternalService("myservice", "10.10.10.10", 156, "myservice.landscape")

	var configObjects []interface{}
	yaml.Unmarshal([]byte(configAsString), &configObjects)

	serviceEntryConfig, err := lookupObjectFromConfig(configObjects, "ServiceEntry")
	g.Expect(err).ShouldNot(HaveOccurred())

	serviceEntrySpec, _ := yaml.Marshal(serviceEntryConfig["spec"])
	serviceEntryMetadata, _ := yaml.Marshal(serviceEntryConfig["metadata"])

	g.Expect(string(serviceEntrySpec)).To(ContainSubstring("10.10.10.10"))
	g.Expect(string(serviceEntrySpec)).To(ContainSubstring("156"))
	g.Expect(string(serviceEntrySpec)).To(ContainSubstring("name: myservice"))
	g.Expect(string(serviceEntryMetadata)).To(ContainSubstring("name: myservice-service-entry"))
}

func TestCompleteVirtualService(t *testing.T) {
	g := NewGomegaWithT(t)

	configAsString, _ := CreateEntriesForExternalService("myservice", "10.10.10.10", 156, "myservice.landscape")

	var configObjects []interface{}
	yaml.Unmarshal([]byte(configAsString), &configObjects)

	virtualServiceConfig, err := lookupObjectFromConfig(configObjects, "VirtualService")
	g.Expect(err).ShouldNot(HaveOccurred())

	virtualServiceSpec, _ := yaml.Marshal(virtualServiceConfig["spec"])

	g.Expect(string(virtualServiceSpec)).To(ContainSubstring("myservice.landscape"))
	g.Expect(string(virtualServiceSpec)).To(ContainSubstring("156"))
	g.Expect(string(virtualServiceSpec)).To(ContainSubstring("host: myservice.service-fabrik"))
}

func lookupObjectFromConfig(configObjects []interface{}, kind string) (map[string]interface{}, error) {

	for _, entryUntyped := range configObjects {
		entry := entryUntyped.(map[string]interface{})
		if entry["kind"] == kind {
			return entry, nil
		}
	}

	return nil, errors.New("not found:" + kind)
}

func TestGatewayFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	gateway := createIngressGatewayForExternalService("postgres.services.cf.dev01.aws.istio.sapcloud.io",
		9000,
		"postgres-server",
		"client.istio.sapcloud.io")

	text, err := toText(model.Gateway, gateway)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(gateway_yml))
}

func TestVirtualServiceFromGo(t *testing.T) {
	g := NewGomegaWithT(t)

	virtualService := createVirtualServiceForExternalService("postgres.services.cf.dev01.aws.istio.sapcloud.io",
		47637,
		"postgres-server")

	text, err := toText(model.VirtualService, virtualService)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(text).To(Equal(virtual_service_yml))
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
