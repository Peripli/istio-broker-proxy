package config

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/gomega"
	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/config/kube/crd"
	"istio.io/istio/pilot/pkg/model"
	"regexp"
	"testing"
)

func TestCompleteEntryNotEmpty(t *testing.T) {
	g := NewGomegaWithT(t)

	configObjects := CreateEntriesForExternalService("myservice", "10.10.10.10", 10, "myservice.landscape")

	g.Expect(configObjects).To(HaveLen(3))
}

func TestCompleteClientEntryNotEmpty(t *testing.T) {
	g := NewGomegaWithT(t)

	configObjects := CreateEntriesForExternalServiceClient("myservice", "myservice.landscape", 1111)

	g.Expect(configObjects).To(HaveLen(7))
}

func TestCompleteEntryGateway(t *testing.T) {
	g := NewGomegaWithT(t)

	configObjects := CreateEntriesForExternalService("myservice", "10.10.10.10", 10, "myservice.landscape")

	gatewaySpec, gatewayMetadata := getSpecAndMetadataFromConfig(g, configObjects, gateway)

	g.Expect(gatewaySpec).To(ContainSubstring("myservice.landscape"))
	g.Expect(gatewaySpec).To(ContainSubstring("9000"))
	g.Expect(gatewaySpec).To(ContainSubstring("client.istio.sapcloud.io"))
	g.Expect(gatewaySpec).To(ContainSubstring("config/certs/myservice.key"))
	g.Expect(gatewaySpec).To(ContainSubstring("config/certs/myservice.crt"))

	g.Expect(gatewayMetadata).To(ContainSubstring("name: myservice-gateway"))
}

func TestCompleteServiceEntry(t *testing.T) {
	g := NewGomegaWithT(t)

	configObjects := CreateEntriesForExternalService("myservice", "10.10.10.10", 156, "myservice.landscape")

	serviceEntrySpec, serviceEntryMetadata := getSpecAndMetadataFromConfig(g, configObjects, serviceEntry)

	g.Expect(serviceEntrySpec).To(ContainSubstring("10.10.10.10"))
	g.Expect(serviceEntrySpec).To(ContainSubstring("156"))
	g.Expect(serviceEntrySpec).To(ContainSubstring("name: myservice"))
	g.Expect(serviceEntryMetadata).To(ContainSubstring("name: myservice-service-entry"))
}

func TestCompleteVirtualService(t *testing.T) {
	g := NewGomegaWithT(t)

	configObjects := CreateEntriesForExternalService("myservice", "10.10.10.10", 156, "myservice.landscape")
	virtualServiceSpec, _ := getSpecAndMetadataFromConfig(g, configObjects, virtualService)

	g.Expect(virtualServiceSpec).To(ContainSubstring("myservice.landscape"))
	g.Expect(virtualServiceSpec).To(ContainSubstring("156"))
	g.Expect(virtualServiceSpec).To(ContainSubstring("host: myservice.service-fabrik"))
}

func TestCompleteEntryClientGateway(t *testing.T) {
	g := NewGomegaWithT(t)

	configObjects := CreateEntriesForExternalServiceClient("myservice", "myservice.landscape", 12345)

	gatewaySpec, gatewayMetadata := getSpecAndMetadataFromConfig(g, configObjects, gateway)

	g.Expect(gatewaySpec).To(ContainSubstring("myservice.landscape"))
	g.Expect(gatewaySpec).To(ContainSubstring("443"))
	g.Expect(gatewaySpec).To(ContainSubstring("spiffe://cluster.local/ns/default/sa/default"))
	g.Expect(gatewaySpec).To(ContainSubstring("/etc/certs/cert-chain.pem"))

	g.Expect(gatewayMetadata).To(ContainSubstring("name: istio-egressgateway-myservice"))
}

func TestCompleteEntryClientEgressDestinationRule(t *testing.T) {
	g := NewGomegaWithT(t)

	configObjects := CreateEntriesForExternalServiceClient("myservice", "myservice.landscape", 12345)

	ruleSpecs, ruleMetadatas := getSpecsAndMetadatasFromConfig(g, configObjects, destinationRule)
	ruleSpec := ruleSpecs[0]
	ruleMetadata := ruleMetadatas[0]

	g.Expect(ruleSpec).To(ContainSubstring("myservice.landscape"))
	g.Expect(ruleSpec).To(ContainSubstring("9000"))
	g.Expect(ruleSpec).To(ContainSubstring("sni: myservice.landscape"))

	g.Expect(ruleMetadata).To(ContainSubstring("egressgateway-myservice"))
}

func TestCompleteEntryClientSidecarDestinationRule(t *testing.T) {
	g := NewGomegaWithT(t)

	configObjects := CreateEntriesForExternalServiceClient("myservice", "myservice.landscape", 12345)

	ruleSpecs, ruleMetadatas := getSpecsAndMetadatasFromConfig(g, configObjects, destinationRule)
	g.Expect(ruleSpecs).To(HaveLen(2))

	ruleSpec := ruleSpecs[1]
	ruleMetadata := ruleMetadatas[1]

	g.Expect(ruleSpec).To(ContainSubstring("istio-egressgateway.istio-system.svc.cluster.local"))
	g.Expect(ruleSpec).To(ContainSubstring("sni: myservice.landscape"))

	g.Expect(ruleMetadata).To(ContainSubstring("sidecar-to-egress-myservice"))
}

func TestCompleteEntryClientServiceEntry(t *testing.T) {
	g := NewGomegaWithT(t)

	configObjects := CreateEntriesForExternalServiceClient("myservice", "myservice.landscape", 12345)

	serviceEntriesConfigs := lookupObjectsFromConfigs(configObjects, serviceEntry)
	g.Expect(serviceEntriesConfigs).To(HaveLen(2))

	var ports []string
	for _, serviceEntryConfig := range serviceEntriesConfigs {
		entrySpec, _ := yaml.Marshal(serviceEntryConfig)
		g.Expect(entrySpec).To(ContainSubstring("myservice.landscape"))

		r := regexp.MustCompile(`number: (\d+)`)
		port := r.FindStringSubmatch(string(entrySpec))[1]
		ports = append(ports, port)
	}
	g.Expect(ports).To(ContainElement("12345"))
	g.Expect(ports).To(ContainElement("9000"))
}

func TestCompleteEntryClientVirtualServices(t *testing.T) {
	g := NewGomegaWithT(t)

	configObjects := CreateEntriesForExternalServiceClient("myservice", "myservice.landscape", 12345)
	serviceSpecs, serviceMetadatas := getSpecsAndMetadatasFromConfig(g, configObjects, virtualService)
	g.Expect(serviceSpecs).To(HaveLen(2))
	g.Expect(serviceMetadatas).To(HaveLen(2))
	g.Expect(serviceSpecs[0]).To(ContainSubstring("mesh"))
	g.Expect(serviceSpecs[1]).To(ContainSubstring("istio-egressgateway-myservice"))
}

func TestCreatesYamlDocuments(t *testing.T) {
	g := NewGomegaWithT(t)
	dummyConfigs := []model.Config{model.Config{Spec: &v1alpha3.ServiceEntry{}}}
	dummyConfigs[0].Type = serviceEntry
	text, err := ToYamlDocuments(dummyConfigs)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(text).To(ContainSubstring("---"))
}

func getSpecAndMetadataFromConfig(g *GomegaWithT, configObjects []model.Config, configType string) (string, string) {
	specs, metadatas := getSpecsAndMetadatasFromConfig(g, configObjects, configType)
	return specs[0], metadatas[0]
}

func getSpecsAndMetadatasFromConfig(g *GomegaWithT, configObjects []model.Config, configType string) ([]string, []string) {
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

func lookupObjectsFromConfigs(configObjects []model.Config, kind string) (array []model.Config) {
	for _, entry := range configObjects {
		if entry.Type == kind {
			array = append(array, entry)
		}
	}

	return array
}
