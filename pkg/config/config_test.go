package config

import (
	"encoding/json"
	"fmt"
	"github.com/Peripli/istio-broker-proxy/pkg/model"
	"io/ioutil"
	"regexp"
	"testing"

	"github.com/ghodss/yaml"
	. "github.com/onsi/gomega"
	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/config/kube/crd"
	istioModel "istio.io/istio/pilot/pkg/model"
)

func TestCompleteEntryNotEmpty(t *testing.T) {
	g := NewGomegaWithT(t)

	configObjects := CreateEntriesForExternalService("myservice", "10.10.10.10", 10, "myservice.landscape", "client.my.client.domain.io", 9000)

	g.Expect(configObjects).To(HaveLen(3))
}

func TestCompleteClientEntryNotEmpty(t *testing.T) {
	g := NewGomegaWithT(t)

	configObjects := CreateEntriesForExternalServiceClient("myservice", "myservice.landscape", "1.1.1.1", 9000, "catalog", "landscape")

	g.Expect(configObjects).To(HaveLen(6))
}

func TestCompleteEntryGateway(t *testing.T) {
	g := NewGomegaWithT(t)

	configObjects := CreateEntriesForExternalService("myservice", "10.10.10.10", 10, "myservice.landscape", "client.my.client.domain.io", 9000)

	gatewaySpec, gatewayMetadata := getSpecAndMetadataFromConfig(g, configObjects, gateway)

	err := istioModel.ValidateGateway("test", "test", configObjects[0].Spec)
	g.Expect(err).NotTo(HaveOccurred())

	err = istioModel.ValidateVirtualService("test", "test", configObjects[1].Spec)
	g.Expect(err).NotTo(HaveOccurred())

	err = istioModel.ValidateServiceEntry("test", "test", configObjects[2].Spec)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(gatewaySpec).To(ContainSubstring("myservice.landscape"))
	g.Expect(gatewaySpec).To(ContainSubstring("9000"))
	g.Expect(gatewaySpec).To(ContainSubstring("client.my.client.domain.io"))
	g.Expect(gatewaySpec).To(ContainSubstring("config/certs/cf-service.key"))
	g.Expect(gatewaySpec).To(ContainSubstring("config/certs/cf-service.crt"))

	g.Expect(gatewayMetadata).To(ContainSubstring("name: myservice-gateway"))
}

func TestCompleteServiceEntry(t *testing.T) {
	g := NewGomegaWithT(t)

	configObjects := CreateEntriesForExternalService("myservice", "10.10.10.10", 156, "myservice.landscape", "client.my.client.domain.io", 9000)

	serviceEntrySpec, serviceEntryMetadata := getSpecAndMetadataFromConfig(g, configObjects, serviceEntry)

	g.Expect(serviceEntrySpec).To(ContainSubstring("10.10.10.10"))
	g.Expect(serviceEntrySpec).To(ContainSubstring("156"))
	g.Expect(serviceEntrySpec).To(ContainSubstring("name: myservice"))
	g.Expect(serviceEntryMetadata).To(ContainSubstring("name: myservice-service-entry"))
}

func TestCompleteVirtualService(t *testing.T) {
	g := NewGomegaWithT(t)

	configObjects := CreateEntriesForExternalService("myservice", "10.10.10.10", 156, "myservice.landscape", "client.my.client.domain.io", 9000)
	virtualServiceSpec, _ := getSpecAndMetadataFromConfig(g, configObjects, virtualService)

	g.Expect(virtualServiceSpec).To(ContainSubstring("myservice.landscape"))
	g.Expect(virtualServiceSpec).To(ContainSubstring("156"))
	g.Expect(virtualServiceSpec).To(ContainSubstring("host: myservice.service-fabrik"))
}

func TestCompleteEntryClientGateway(t *testing.T) {
	g := NewGomegaWithT(t)

	configObjects := CreateEntriesForExternalServiceClient("myservice", "myservice.landscape", "1.1.1.1", 9000, "catalog", "landscape")

	gatewaySpec, gatewayMetadata := getSpecAndMetadataFromConfig(g, configObjects, "gateway")

	g.Expect(gatewaySpec).To(ContainSubstring("myservice.landscape"))
	g.Expect(gatewaySpec).To(ContainSubstring("443"))
	g.Expect(gatewaySpec).To(ContainSubstring("spiffe://cluster.local/ns/default/sa/default"))
	g.Expect(gatewaySpec).To(ContainSubstring("/etc/certs/cert-chain.pem"))

	g.Expect(gatewayMetadata).To(ContainSubstring("name: istio-egressgateway-myservice"))
	g.Expect(gatewayMetadata).To(ContainSubstring("service: myservice"))
}

func TestCompleteEntryClientEgressDestinationRule(t *testing.T) {
	g := NewGomegaWithT(t)

	configObjects := CreateEntriesForExternalServiceClient("myservice", "myservice.landscape", "1.1.1.1", 9000, "catalog", "landscape")

	ruleSpecs, ruleMetadatas := getSpecsAndMetadatasFromConfig(g, configObjects, istio_destinationRule)
	ruleSpec := ruleSpecs[0]
	ruleMetadata := ruleMetadatas[0]

	g.Expect(ruleSpec).To(ContainSubstring("myservice.landscape"))
	g.Expect(ruleSpec).To(ContainSubstring("9000"))
	g.Expect(ruleSpec).To(ContainSubstring("sni: myservice.landscape"))

	g.Expect(ruleMetadata).To(ContainSubstring("egressgateway-myservice"))
}

func TestCompleteEntryClientSidecarDestinationRule(t *testing.T) {
	g := NewGomegaWithT(t)

	configObjects := CreateEntriesForExternalServiceClient("myservice", "myservice.landscape", "1.1.1.1", 9000, "catalog", "landscape")

	ruleSpecs, ruleMetadatas := getSpecsAndMetadatasFromConfig(g, configObjects, istio_destinationRule)
	g.Expect(ruleSpecs).To(HaveLen(2))

	ruleSpec := ruleSpecs[1]
	ruleMetadata := ruleMetadatas[1]

	g.Expect(ruleSpec).To(ContainSubstring("istio-egressgateway.istio-system.svc.cluster.local"))
	g.Expect(ruleSpec).To(ContainSubstring("sni: myservice.landscape"))

	g.Expect(ruleMetadata).To(ContainSubstring("sidecar-to-egress-myservice"))
}

func TestCompleteEntryClientServiceEntry(t *testing.T) {
	g := NewGomegaWithT(t)

	configObjects := CreateEntriesForExternalServiceClient("myservice", "myservice.landscape", "1.1.1.1", 9000, "catalog", "landscape")

	serviceEntriesConfigs := lookupObjectsFromConfigs(configObjects, istio_serviceEntry)
	g.Expect(serviceEntriesConfigs).To(HaveLen(1))

	var ports []string
	for _, serviceEntryConfig := range serviceEntriesConfigs {
		entrySpec, _ := yaml.Marshal(serviceEntryConfig)
		g.Expect(entrySpec).To(ContainSubstring("myservice.landscape"))

		r := regexp.MustCompile(`number: (\d+)`)
		port := r.FindStringSubmatch(string(entrySpec))[1]
		ports = append(ports, port)
	}
	g.Expect(ports).To(ContainElement("9000"))
}

func TestCompleteEntryClientVirtualServices(t *testing.T) {
	g := NewGomegaWithT(t)

	configObjects := CreateEntriesForExternalServiceClient("myservice", "myservice.landscape", "1.1.1.1", 9000, "catalog", "landscape")
	serviceSpecs, serviceMetadatas := getSpecsAndMetadatasFromConfig(g, configObjects, istio_virtualService)
	g.Expect(serviceSpecs).To(HaveLen(2))
	g.Expect(serviceMetadatas).To(HaveLen(2))
	g.Expect(serviceSpecs[0]).To(ContainSubstring("mesh"))
	g.Expect(serviceSpecs[1]).To(ContainSubstring("istio-egressgateway-myservice"))
}

func TestCreatesYamlDocuments(t *testing.T) {
	g := NewGomegaWithT(t)
	dummyConfigs := []istioModel.Config{{Spec: &v1alpha3.ServiceEntry{}}}
	dummyConfigs[0].Type = serviceEntry
	text, err := ToYamlDocuments(dummyConfigs)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(text).To(ContainSubstring("---"))
}

func TestErrorInToText(t *testing.T) {
	g := NewGomegaWithT(t)

	_, err := enrichAndtoText(istioModel.Config{})

	g.Expect(err).Should(HaveOccurred())
}

func TestCreateIstioConfigForProvider(t *testing.T) {
	g := NewGomegaWithT(t)

	eps := []model.Endpoint{{"1.1.1.1", 9999}, {"2.2.2.2", 7777}}

	request := model.BindRequest{
		NetworkData: model.NetworkDataRequest{
			NetworkProfileId: "my-profile-id",
			Data:             model.DataRequest{ConsumerId: "147"}}}
	response := model.BindResponse{
		NetworkData: model.NetworkDataResponse{
			NetworkProfileId: "your-profile-id",
			Data: model.DataResponse{
				ProviderId: "852",
			}},
		Endpoints: eps,

		Credentials: model.Credentials{AdditionalProperties: map[string]json.RawMessage{"user": json.RawMessage([]byte(`"myuser"`))}}}

	istioConfig := CreateIstioConfigForProvider(&request, &response, "my-binding-id", "services.my-domain")

	gatewaySpec, gatewayMetadata := getSpecsAndMetadatasFromConfig(g, istioConfig, gateway)

	virtualServiceSpec, _ := getSpecsAndMetadatasFromConfig(g, istioConfig, virtualService)

	serviceEntrySpec, _ := getSpecsAndMetadatasFromConfig(g, istioConfig, serviceEntry)

	g.Expect(len(gatewaySpec)).To(Equal(len(eps)))
	g.Expect(len(gatewayMetadata)).To(Equal(len(eps)))
	g.Expect(len(virtualServiceSpec)).To(Equal(len(eps)))
	g.Expect(len(serviceEntrySpec)).To(Equal(len(eps)))

	err := istioModel.ValidateGateway("test", "test", istioConfig[0].Spec)
	g.Expect(err).NotTo(HaveOccurred())

	err = istioModel.ValidateVirtualService("test", "test", istioConfig[1].Spec)
	g.Expect(err).NotTo(HaveOccurred())

	err = istioModel.ValidateServiceEntry("test", "test", istioConfig[2].Spec)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(gatewaySpec[0]).To(ContainSubstring("147"))
	g.Expect(gatewaySpec[0]).To(ContainSubstring("0.my-binding-id.services.my-domain"))
	g.Expect(gatewayMetadata[0]).To(ContainSubstring("name: 0-my-binding-id-gateway"))
	g.Expect(serviceEntrySpec[0]).To(ContainSubstring("- address: 1.1.1.1"))
	g.Expect(serviceEntrySpec[1]).To(ContainSubstring("- address: 2.2.2.2"))
	g.Expect(virtualServiceSpec[0]).To(ContainSubstring("number: 9999"))
	g.Expect(virtualServiceSpec[1]).To(ContainSubstring("number: 7777"))

}

func TestValidIdentifier(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(createValidIdentifer("10.1.2.3")).To(Equal("10-1-2-3"))
	g.Expect(createValidIdentifer(".10.1.2.3")).To(Equal("10-1-2-3"))
	g.Expect(createValidIdentifer("&10^1%2$3")).To(Equal("10-1-2-3"))
	g.Expect(createValidIdentifer("ABC-123")).To(Equal("abc-123"))

}

func TestCreateManyServerConfigs(t *testing.T) {
	t.Skip("Only for mass testing")
	g := NewGomegaWithT(t)
	istioConfig := make([]istioModel.Config, 0)

	for i := 0; i < 5000; i++ {
		serviceName := fmt.Sprintf("servicename%d", i)
		endpointServiceEntry := fmt.Sprintf("192.168.%d.%d", i/256, i%256)
		port := uint32(i + 10000)
		hostVirtualService := fmt.Sprintf("hostvirtualservice%d.nonsense.de", i)
		istioConfig = append(istioConfig, CreateEntriesForExternalService(serviceName, endpointServiceEntry, port, hostVirtualService, "test.sap.local", 9000)...)
	}

	content, err := ToYamlDocuments(istioConfig)
	g.Expect(err).NotTo(HaveOccurred())
	err = ioutil.WriteFile("test5000.yml", []byte(content), 0775)
	g.Expect(err).NotTo(HaveOccurred())
}

func getSpecAndMetadataFromConfig(g *GomegaWithT, configObjects []istioModel.Config, configType string) (string, string) {
	specs, metadatas := getSpecsAndMetadatasFromConfig(g, configObjects, configType)

	g.Expect(specs).NotTo(HaveLen(0), "object not found: %s", configType)
	return specs[0], metadatas[0]
}

func getSpecsAndMetadatasFromConfig(g *GomegaWithT, configObjects []istioModel.Config, configType string) ([]string, []string) {
	configs := lookupObjectsFromConfigs(configObjects, configType)
	var specs, metadatas []string
	for _, config := range configs {
		kubernetesConf, err := crd.ConvertConfig(schemas[config.Type], config)
		g.Expect(err).ShouldNot(HaveOccurred())
		spec, err := yaml.Marshal(kubernetesConf.GetSpec())
		g.Expect(err).ShouldNot(HaveOccurred())
		specs = append(specs, string(spec))
		metadata, err := yaml.Marshal(kubernetesConf.GetObjectMeta())
		g.Expect(err).ShouldNot(HaveOccurred())
		metadatas = append(metadatas, string(metadata))
	}
	return specs, metadatas
}

func lookupObjectsFromConfigs(configObjects []istioModel.Config, kind string) (array []istioModel.Config) {
	for _, entry := range configObjects {
		if entry.Type == kind {
			array = append(array, entry)
		}
	}

	return array
}
