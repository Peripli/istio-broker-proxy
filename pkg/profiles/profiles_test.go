package profiles

import (
	"encoding/json"
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
	Port int    `json:"port"`
}

func TestAddIstioNetworkDataHasConfigurableProviderId(t *testing.T) {
	g := NewGomegaWithT(t)

	addIstioDataFunc := AddIstioNetworkDataToResponse("my-provider", "", "", 0)

	body := []byte(`{"something_else": "body of response", "endpoints": []}`)
	bodyWithIstioData, err := addIstioDataFunc(body)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(bodyWithIstioData).NotTo(BeNil())

	parsedBody := bodyData{}
	json.Unmarshal(bodyWithIstioData, &parsedBody)
	g.Expect(parsedBody.NetworkData.NetworkProfileId).To(Equal("urn:com.sap.istio:public"))
	g.Expect(parsedBody.NetworkData.ProviderId).To(Equal("my-provider"))
	g.Expect(parsedBody.SomethingElse).To(Equal("body of response"))
}

func TestCreateEndpointHosts(t *testing.T) {
	g := NewGomegaWithT(t)

	serviceId := "postgres-34de6ac"
	systemDomain := "istio.sapcloud.io"
	endpoints := make([]interface{}, 2)

	endpointHost, err := createEndpointHostsBasedOnSystemDomainServiceId(serviceId, systemDomain, endpoints)

	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(endpointHost).To(HaveLen(2))
	g.Expect(endpointHost).To(ContainElement("1.postgres-34de6ac.istio.sapcloud.io"))
	g.Expect(endpointHost).To(ContainElement("2.postgres-34de6ac.istio.sapcloud.io"))
}

func TestAddIstioNetworkDataProvidesEndpointHosts(t *testing.T) {
	g := NewGomegaWithT(t)

	addIstioDataFunc := AddIstioNetworkDataToResponse("my-provider", "postgres-34de6ac", "istio.sapcloud.io", 9000)

	body := []byte(`{"something_else": "body of response", "endpoints": [{}, {}]}`)
	bodyWithIstioData, err := addIstioDataFunc(body)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(bodyWithIstioData).NotTo(BeNil())

	parsedBody := bodyData{}
	json.Unmarshal(bodyWithIstioData, &parsedBody)
	g.Expect(parsedBody.NetworkData.Endpoints).NotTo(BeNil())
	g.Expect(parsedBody.NetworkData.Endpoints).To(HaveLen(2))
	g.Expect(parsedBody.NetworkData.Endpoints[0].Host).To(ContainSubstring("1.postgres-34de6ac.istio.sapcloud.io"))
	g.Expect(parsedBody.NetworkData.Endpoints[0].Port).To(Equal(9000))
}
