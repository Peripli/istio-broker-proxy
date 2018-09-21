package model

import (
	"encoding/json"
	. "github.com/onsi/gomega"
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

func TestBindRequestUnmarshalInvalidAdditionalProperties(t *testing.T) {
	g := NewGomegaWithT(t)
	var bindRequest BindRequest
	err := json.Unmarshal([]byte(`[]`), &bindRequest)
	g.Expect(err).To(HaveOccurred())
}

func TestBindRequestUnmarshalInvalidNetworkData(t *testing.T) {
	g := NewGomegaWithT(t)
	var bindRequest BindRequest
	err := json.Unmarshal([]byte(`{
		"network_data": 666
    }`), &bindRequest)
	g.Expect(err).To(HaveOccurred())
}
