package profiles

import (
	"encoding/json"
	. "github.com/onsi/gomega"

	//"os"
	"testing"
)

type requestBodyData struct {
	SomethingElse string             `json:"something_else,omitempty"`
	NetworkData   requestNetworkData `json:"network_data"`
}

type requestNetworkData struct {
	NetworkProfileId string `json:"network_profile_id"`
	ConsumerId       string `json:"consumer_id"`
}

type responseBodyData struct {
	SomethingElse string              `json:"something_else,omitempty"`
	NetworkData   responseNetworkData `json:"network_data"`
}

type responseNetworkData struct {
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

	body := []byte(`{"something_else": "body of response", "endpoints": [{}]}`)
	bodyWithIstioData, err := addIstioDataFunc(body)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(bodyWithIstioData).NotTo(BeNil())

	parsedBody := responseBodyData{}
	json.Unmarshal(bodyWithIstioData, &parsedBody)
	g.Expect(parsedBody.NetworkData.NetworkProfileId).To(Equal("urn:com.sap.istio:public"))
	g.Expect(parsedBody.NetworkData.ProviderId).To(Equal("my-provider"))
	g.Expect(parsedBody.SomethingElse).To(Equal("body of response"))
}

func TestAddIstioNetworkDataWithInvalidEndpoints(t *testing.T) {
	g := NewGomegaWithT(t)

	addIstioDataFunc := AddIstioNetworkDataToResponse("my-provider", "", "", 0)

	body := []byte(`{"something_else": "body of response", "endpoints": {"noarray": true}}`)
	_, err := addIstioDataFunc(body)

	g.Expect(err).To(HaveOccurred())
}

func TestAddIstioNetworkDataWithInvalidJson(t *testing.T) {
	g := NewGomegaWithT(t)

	addIstioDataFunc := AddIstioNetworkDataToResponse("my-provider", "", "", 0)

	body := []byte(`{"something_invalid}`)
	_, err := addIstioDataFunc(body)

	g.Expect(err).To(HaveOccurred())
}

func TestCreateEndpointHosts(t *testing.T) {
	g := NewGomegaWithT(t)

	serviceId := "postgres-34de6ac"
	systemDomain := "istio.sapcloud.io"

	endpointHost := createEndpointHostsBasedOnSystemDomainServiceId(serviceId, systemDomain, 2)

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

	parsedBody := responseBodyData{}
	json.Unmarshal(bodyWithIstioData, &parsedBody)
	g.Expect(parsedBody.NetworkData.Endpoints).NotTo(BeNil())
	g.Expect(parsedBody.NetworkData.Endpoints).To(HaveLen(2))
	g.Expect(parsedBody.NetworkData.Endpoints[0].Host).To(ContainSubstring("1.postgres-34de6ac.istio.sapcloud.io"))
	g.Expect(parsedBody.NetworkData.Endpoints[0].Port).To(Equal(9000))
}

func TestBlueprintServiceDoesntCrash(t *testing.T) {
	g := NewGomegaWithT(t)
	addIstioDataFunc := AddIstioNetworkDataToResponse("my-provider", "postgres-34de6ac", "istio.sapcloud.io", 9000)
	body := []byte(`{"credentials":{"hosts":["10.11.31.128"],"hostname":"10.11.31.128","port":8080,"uri":"http://50da4fff492a97c635a4bfe4fc64276e:160bbfd6e913f353e6f4ea526e8e58df@10.11.31.128:8080","username":"50da4fff492a97c635a4bfe4fc64276e","password":"160bbfd6e913f353e6f4ea526e8e58df"}}`)
	resultBody, err := addIstioDataFunc(body)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(resultBody).To(Equal(body))
}

func TestAddIstioNetworkToRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	addIstioDataFunc := AddIstioNetworkDataToRequest("my-consumer")

	body := []byte(`{"something_else": "body of response"}`)
	bodyWithIstioData, err := addIstioDataFunc(body)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(bodyWithIstioData).NotTo(BeNil())

	parsedBody := requestBodyData{}
	json.Unmarshal(bodyWithIstioData, &parsedBody)
	g.Expect(parsedBody.NetworkData.ConsumerId).To(Equal("my-consumer"))
}
