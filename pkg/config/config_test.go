package config

import (
	"errors"
	"github.com/ghodss/yaml"
	. "github.com/onsi/gomega"
	"testing"
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
