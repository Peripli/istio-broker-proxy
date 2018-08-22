package config

import (
	"errors"
	"github.com/ghodss/yaml"
	. "github.com/onsi/gomega"
	"regexp"
	"testing"
)

func TestCompleteEntryNotEmpty(t *testing.T) {
	g := NewGomegaWithT(t)

	configAsString, _ := CreateEntriesForExternalService("myservice", "10.10.10.10", 10, "myservice.landscape")

	var configObjects []interface{}
	yaml.Unmarshal([]byte(configAsString), &configObjects)

	g.Expect(configObjects).To(HaveLen(3))
}

func TestCompleteClientEntryNotEmpty(t *testing.T) {
	g := NewGomegaWithT(t)

	configAsString, _ := CreateEntriesForExternalServiceClient("myservice", "myservice.landscape", 1111)

	var configObjects []interface{}
	yaml.Unmarshal([]byte(configAsString), &configObjects)

	g.Expect(configObjects).To(HaveLen(6))
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

func TestCompleteEntryClientGateway(t *testing.T) {
	g := NewGomegaWithT(t)

	configAsString, _ := CreateEntriesForExternalServiceClient("myservice", "myservice.landscape", 12345)

	var configObjects []interface{}
	yaml.Unmarshal([]byte(configAsString), &configObjects)

	gatewayConfig, err := lookupObjectFromConfig(configObjects, "Gateway")
	g.Expect(err).ShouldNot(HaveOccurred())

	gatewaySpec, _ := yaml.Marshal(gatewayConfig["spec"])
	gatewayMetadata, _ := yaml.Marshal(gatewayConfig["metadata"])

	g.Expect(string(gatewaySpec)).To(ContainSubstring("myservice.landscape"))
	g.Expect(string(gatewaySpec)).To(ContainSubstring("443"))
	g.Expect(string(gatewaySpec)).To(ContainSubstring("spiffe://cluster.local/ns/default/sa/default"))
	g.Expect(string(gatewaySpec)).To(ContainSubstring("/etc/certs/cert-chain.pem"))

	g.Expect(string(gatewayMetadata)).To(ContainSubstring("name: istio-egressgateway-myservice"))
}

func TestCompleteEntryClientDestinationRule(t *testing.T) {
	g := NewGomegaWithT(t)

	configAsString, _ := CreateEntriesForExternalServiceClient("myservice", "myservice.landscape", 12345)

	var configObjects []interface{}
	yaml.Unmarshal([]byte(configAsString), &configObjects)

	ruleConfig, err := lookupObjectFromConfig(configObjects, "DestinationRule")
	g.Expect(err).ShouldNot(HaveOccurred())

	ruleSpec, _ := yaml.Marshal(ruleConfig["spec"])
	ruleMetadata, _ := yaml.Marshal(ruleConfig["metadata"])

	g.Expect(string(ruleSpec)).To(ContainSubstring("myservice.landscape"))
	g.Expect(string(ruleSpec)).To(ContainSubstring("9000"))
	g.Expect(string(ruleSpec)).To(ContainSubstring("sni: myservice.landscape"))

	g.Expect(string(ruleMetadata)).To(ContainSubstring("egressgateway-myservice"))
}

func TestCompleteEntryClientServiceEntry(t *testing.T) {
	g := NewGomegaWithT(t)

	configAsString, _ := CreateEntriesForExternalServiceClient("myservice", "myservice.landscape", 12345)

	var configObjects []interface{}
	yaml.Unmarshal([]byte(configAsString), &configObjects)

	serviceEntriesConfigs := lookupObjectsFromConfigs(configObjects, "ServiceEntry")
	g.Expect(serviceEntriesConfigs).To(HaveLen(2))

	var ports []string
	for _, serviceEntryConfig := range serviceEntriesConfigs {
		entrySpec, _ := yaml.Marshal(serviceEntryConfig)
		g.Expect(string(entrySpec)).To(ContainSubstring("myservice.landscape"))

		r := regexp.MustCompile(`number: (\d+)`)
		port := r.FindStringSubmatch(string(entrySpec))[1]
		ports = append(ports, port)
	}
	g.Expect(ports).To(ContainElement("12345"))
	g.Expect(ports).To(ContainElement("9000"))
}

func lookupObjectFromConfig(configObjects []interface{}, kind string) (map[string]interface{}, error) {

	array := lookupObjectsFromConfigs(configObjects, kind)

	if array == nil {

		return nil, errors.New("not found:" + kind)
	} else {
		return array[0], nil
	}
}

func lookupObjectsFromConfigs(configObjects []interface{}, kind string) []map[string]interface{} {
	var array []map[string]interface{}

	for _, entryUntyped := range configObjects {
		entry := entryUntyped.(map[string]interface{})
		if entry["kind"] == kind {
			array = append(array, entry)
		}
	}

	return array
}
