package profiles

import (
	"encoding/json"
	"fmt"
	. "github.com/onsi/gomega"

	//"os"
	"testing"
)

type bodyData struct {
	SomethingElse string      `json:"something_else,omitempty"`
	NetworkData   networkData `json:"network_data"`
}

type networkData struct {
	NetworkProfileId string      `json:"network_profile_id"`
	ProviderId       string      `json:"provider_id"`
	Endpoints        []endpoints `json:"endpoints, omitempty"`
}

type endpoints struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

func TestAddIstioNetworkData(t *testing.T) {
	g := NewGomegaWithT(t)

	body := []byte(`{}`)
	bodyWithIstioData, err := AddIstioNetworkData(body)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(bodyWithIstioData).NotTo(BeNil())

	parsedBody := bodyData{}
	json.Unmarshal(bodyWithIstioData, &parsedBody)
	g.Expect(parsedBody.NetworkData.NetworkProfileId).To(Equal("urn:com.sap.istio:public"))
	g.Expect(parsedBody.NetworkData.ProviderId).To(Equal("my-provider"))
}

func TestAddIstioNetworkDataPreservesOriginalBody(t *testing.T) {
	g := NewGomegaWithT(t)

	body := []byte(`{"something_else": "body of response"}`)
	bodyWithIstioData, err := AddIstioNetworkData(body)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(bodyWithIstioData).NotTo(BeNil())

	parsedBody := bodyData{}
	json.Unmarshal(bodyWithIstioData, &parsedBody)
	g.Expect(parsedBody.NetworkData.NetworkProfileId).To(Equal("urn:com.sap.istio:public"))
	g.Expect(parsedBody.NetworkData.ProviderId).To(Equal("my-provider"))
	g.Expect(parsedBody.SomethingElse).To(Equal("body of response"))
}

func TestAddIstioNetworkDataHasConfigurableProviderId(t *testing.T) {
	g := NewGomegaWithT(t)

	addIstioDataFunc := CreateConfigurableNetworkProfile("my-provider", "", "", nil)

	body := []byte(`{"something_else": "body of response"}`)
	bodyWithIstioData, err := addIstioDataFunc(body)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(bodyWithIstioData).NotTo(BeNil())

	parsedBody := bodyData{}
	json.Unmarshal(bodyWithIstioData, &parsedBody)
	g.Expect(parsedBody.NetworkData.NetworkProfileId).To(Equal("urn:com.sap.istio:public"))
	g.Expect(parsedBody.NetworkData.ProviderId).To(Equal("my-provider"))
	g.Expect(parsedBody.SomethingElse).To(Equal("body of response"))
}

func TestAddIstioNetworkDataProvidesEndpointHosts(t *testing.T) {
	g := NewGomegaWithT(t)

	serviceId := "postgres-34de6ac"
	systemDomain := "istio.sapcloud.io"
	endpoints := make([]interface{}, 2)

	endpointHost, err := createEndpointHostsBasedOnSystemDomainServiceId(serviceId, systemDomain, endpoints)

	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(endpointHost).To(HaveLen(2))
	g.Expect(endpointHost).To(ContainElement("1.postgres-34de6ac.istio.sapcloud.io"))
	g.Expect(endpointHost).To(ContainElement("2.postgres-34de6ac.istio.sapcloud.io"))

	addIstioDataFunc := CreateConfigurableNetworkProfile("my-provider", serviceId, systemDomain, endpoints)

	body := []byte(`{"something_else": "body of response"}`)
	bodyWithIstioData, err := addIstioDataFunc(body)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(bodyWithIstioData).NotTo(BeNil())

	g.Expect(string(bodyWithIstioData)).To(ContainSubstring("2.postgres-34de6ac.istio.sapcloud.io"))

	parsedBody := bodyData{}
	json.Unmarshal(bodyWithIstioData, &parsedBody)
	fmt.Printf("ep: %v", parsedBody.NetworkData.Endpoints)
	g.Expect(parsedBody.NetworkData.Endpoints).NotTo(BeNil())
	//ToDo why are additional array elements not parsed?
	g.Expect(parsedBody.NetworkData.Endpoints).To(HaveLen(2))
	g.Expect(parsedBody.NetworkData.Endpoints[0].Host).To(ContainSubstring("1.postgres-34de6ac.istio.sapcloud.io"))

}

//func TestAddIstioNetworkDataProvidesEndpointHostsBasedOnSystemDomainServiceIdAndEndpointIndex(t *testing.T) {
//	g := NewGomegaWithT(t)
//
//	//e.g.
//	// serviceId = "postgres-34de6ac"
//	// systemDomain = "istio.sapcloud.io"
//	// two endpoints
//	// 1.postgres-34de6ac.istio.sapcloud.io
//	// 2.postgres-34de6ac.istio.sapcloud.io
//
//	//serviceId can be found via ctx.Params.GetByName instance_id
//	//systemdomain via global config
//	//endpoints via upstreambroker (response)
//
//	g.Expect(true).To(BeFalse())
//}

func TestAddIstioNetworkDataHasConfigurableEndpointPort(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(true).To(BeFalse())
}
