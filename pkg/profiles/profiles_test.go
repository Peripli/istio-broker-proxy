package profiles

import (
	"encoding/json"
	. "github.com/onsi/gomega"
	//"os"
	"testing"
)

func TestAddIstioNetworkDataHasConfigurableProviderId(t *testing.T) {
	g := NewGomegaWithT(t)

	var body BindResponse
	json.Unmarshal([]byte(`{"something_else": "body of response", "endpoints": [{}]}`), &body)
	AddIstioNetworkDataToResponse("my-provider", "", "", 0, &body)

	g.Expect(body).NotTo(BeNil())

	g.Expect(body.NetworkData.NetworkProfileId).To(Equal("urn:com.sap.istio:public"))
	g.Expect(body.NetworkData.Data.ProviderId).To(Equal("my-provider"))
	g.Expect(string(body.AdditionalProperties["something_else"])).To(Equal(`"body of response"`))
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

	var body BindResponse
	json.Unmarshal([]byte(`{"something_else": "body of response", "endpoints": [{}, {}]}`), &body)
	AddIstioNetworkDataToResponse("my-provider", "postgres-34de6ac", "istio.sapcloud.io", 9000, &body)

	g.Expect(body).NotTo(BeNil())

	g.Expect(body.NetworkData.Data.Endpoints).NotTo(BeNil())
	g.Expect(body.NetworkData.Data.Endpoints).To(HaveLen(2))
	g.Expect(body.NetworkData.Data.Endpoints[0].Host).To(ContainSubstring("1.postgres-34de6ac.istio.sapcloud.io"))
	g.Expect(body.NetworkData.Data.Endpoints[0].Port).To(Equal(9000))
}

func TestBlueprintServiceDoesntCrash(t *testing.T) {
	g := NewGomegaWithT(t)
	compareBody :=
		[]byte(`{"credentials":{"hosts":["10.11.31.128"],"hostname":"10.11.31.128","port":8080,"uri":"http://50da4fff492a97c635a4bfe4fc64276e:160bbfd6e913f353e6f4ea526e8e58df@10.11.31.128:8080","username":"50da4fff492a97c635a4bfe4fc64276e","password":"160bbfd6e913f353e6f4ea526e8e58df"}, "network_data": {
                "network_profile_id": "urn:com.sap.istio:public", "data": { 
                   "provider_id": "my-provider",
                  "endpoints": []
                }
              }}`)
	var bindResponse BindResponse
	json.Unmarshal(compareBody, &bindResponse)
	AddIstioNetworkDataToResponse("my-provider", "postgres-34de6ac", "istio.sapcloud.io", 9000, &bindResponse)
	body, err := json.Marshal(bindResponse)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(body)).To(MatchJSON(compareBody))
}
