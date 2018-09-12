package profiles

import (
	"encoding/json"
	. "github.com/onsi/gomega"
	endpoints "github.infra.hana.ondemand.com/istio/istio-broker/pkg/endpoints"
	"testing"
)

func TestBindRequestUnmarshal(t *testing.T) {
	g := NewGomegaWithT(t)
	var bindRequest BindRequest
	err := json.Unmarshal([]byte(`{
       "abc" : "1234",
		"network_data": {
			"network_profile_id" : "my-profile-id",
			"data":{
				"consumer_id": "147"
			}
		}
    }`), &bindRequest)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(bindRequest.AdditionalProperties["abc"])).To(Equal(`"1234"`))
	g.Expect(string(bindRequest.NetworkData.NetworkProfileId)).To(Equal("my-profile-id"))
	g.Expect(string(bindRequest.NetworkData.Data.ConsumerId)).To(Equal("147"))
}

func TestBindRequestMarshal(t *testing.T) {
	g := NewGomegaWithT(t)
	var bindRequest = BindRequest{
		AdditionalProperties: map[string]json.RawMessage{
			"abc": json.RawMessage([]byte(`"1234"`)),
		},
		NetworkData: NetworkDataRequest{
			NetworkProfileId: "my-profile-id",
			Data:             DataRequest{ConsumerId: "147"}}}
	body, err := json.Marshal(bindRequest)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(body)).To(MatchJSON(`{
        "abc" : "1234",
		"network_data": {
			"network_profile_id" : "my-profile-id",
			"data":{
				"consumer_id": "147"
			}
		}
    }`))
}

func TestBindResponseUnmarshal(t *testing.T) {
	g := NewGomegaWithT(t)
	var bindResponse BindResponse
	err := json.Unmarshal([]byte(`{
       "abc" : "1234",
		"network_data": {
			"network_profile_id" : "your-profile-id",
			"data":{
				"provider_id": "852",
				"endpoints":[
				{"host":"host1", "port": 9999},
				{"host":"host2", "port": 7777}
				]
			}
		},
		"credentials": { "user" : "myuser" }
    }`), &bindResponse)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(bindResponse.AdditionalProperties["abc"])).To(Equal(`"1234"`))
	g.Expect(bindResponse.NetworkData.NetworkProfileId).To(Equal("your-profile-id"))
	g.Expect(bindResponse.NetworkData.Data.ProviderId).To(Equal("852"))
	g.Expect(bindResponse.NetworkData.Data.Endpoints[0]).To(Equal(endpoints.Endpoint{"host1", 9999}))
	g.Expect(string(bindResponse.Credentials)).To(MatchJSON(`{ "user" : "myuser" }`))

}

func TestBindResponseMarshal(t *testing.T) {
	g := NewGomegaWithT(t)
	bindResponse := BindResponse{
		AdditionalProperties: map[string]json.RawMessage{
			"abc": json.RawMessage([]byte(`"1234"`)),
		},
		NetworkData: NetworkDataResponse{
			NetworkProfileId: "your-profile-id",
			Data: DataResponse{
				ProviderId: "852",
				Endpoints:  []endpoints.Endpoint{{"host1", 9999}, {"host2", 7777}},
			}},
		Credentials: json.RawMessage([]byte(`{"user": "myuser"}`))}
	body, err := json.Marshal(bindResponse)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(body)).To(MatchJSON(`{
       "abc" : "1234",
		"network_data": {
			"network_profile_id" : "your-profile-id",
			"data":{
				"provider_id": "852",
				"endpoints":[
				{"host":"host1", "port": 9999},
				{"host":"host2", "port": 7777}
				]
			}
		},
		"credentials": { "user" : "myuser" }
    }`))
}
