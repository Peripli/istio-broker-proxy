package config

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/gomega"
	"istio.io/api/networking/v1alpha3"
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
func createServiceEntryForExternalService(endpointAddress string, port uint32, serviceName string) model.Config {
	hosts := []string{serviceName + ".service-fabrik"}
	ports := v1alpha3.Port{Number: port, Name: "postgres", Protocol: "TCP"}
	resolution := v1alpha3.ServiceEntry_STATIC
	endpoint := v1alpha3.ServiceEntry_Endpoint{Address: endpointAddress}
	serviceEntry := v1alpha3.ServiceEntry{Hosts: hosts, Ports: []*v1alpha3.Port{&ports}, Resolution: resolution,
		Endpoints: []*v1alpha3.ServiceEntry_Endpoint{&endpoint}}
	config := model.Config{Spec: &serviceEntry}
	config.Name = serviceName + "-service-entry"

	return config
}

func createVirtualServiceForExternalService(hostName string, port uint32, serviceName string) model.Config {
	destination := v1alpha3.Destination{Host: serviceName + ".service-fabrik",
		Port: &v1alpha3.PortSelector{Port: &v1alpha3.PortSelector_Number{Number: port}}}
	route := v1alpha3.TCPRoute{Route: []*v1alpha3.DestinationWeight{&v1alpha3.DestinationWeight{Destination: &destination}}}
	tcpRoutes := []*v1alpha3.TCPRoute{&route}
	hosts := []string{hostName}
	gateways := []string{serviceName + "-gateway"}
	virtualService := v1alpha3.VirtualService{Tcp: tcpRoutes, Hosts: hosts, Gateways: gateways}
	config := model.Config{Spec: &virtualService}
	config.Name = serviceName + "-virtual-service"

	return config
}

func toText(schema model.ProtoSchema, config model.Config) (string, error) {
	kubernetesConf, err := crd.ConvertConfig(schema, config)
	if err != nil {
		return "", err
	}
	bytes, err := yaml.Marshal(kubernetesConf)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func createIngressGatewayForExternalService(hostName string, portNumber uint32, serviceName string, clientName string) model.Config {
	port := v1alpha3.Port{Number: portNumber, Name: "tls", Protocol: "TLS"}
	hosts := []string{hostName}
	tls := v1alpha3.Server_TLSOptions{Mode: v1alpha3.Server_TLSOptions_MUTUAL,
		ServerCertificate: "/var/vcap/jobs/envoy/config/certs/" + serviceName + ".crt",
		PrivateKey:        "/var/vcap/jobs/envoy/config/certs/" + serviceName + ".key",
		CaCertificates:    "/var/vcap/jobs/envoy/config/certs/ca.crt",
		SubjectAltNames:   []string{clientName}}
	gateway := v1alpha3.Gateway{Servers: []*v1alpha3.Server{&v1alpha3.Server{Port: &port, Hosts: hosts, Tls: &tls}}}
	config := model.Config{Spec: &gateway}
	config.Name = serviceName + "-gateway"

	return config
}
