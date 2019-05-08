package model

import (
	"encoding/json"
	. "github.com/onsi/gomega"
	"testing"
)

func TestProvisionResponseUnmarshal(t *testing.T) {
	g := NewGomegaWithT(t)
	var provisionResponse ProvisionResponse
	err := json.Unmarshal([]byte(`{
       "abc" : "1234",
		"network_profiles": [{
			"id" : "my-profile-id",
			"data":{
				"consumer_id": "147"
			}
		}]
    }`), &provisionResponse)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(provisionResponse.AdditionalProperties["abc"])).To(Equal(`"1234"`))
	g.Expect(string(provisionResponse.NetworkProfiles[0].ID)).To(Equal("my-profile-id"))
	g.Expect(string(provisionResponse.NetworkProfiles[0].Data)).To(MatchJSON(`{"consumer_id": "147"}`))
}

func TestProvisionResponseMarshal(t *testing.T) {
	g := NewGomegaWithT(t)
	var provisionResponse = ProvisionResponse{
		AdditionalProperties: map[string]json.RawMessage{
			"abc": json.RawMessage([]byte(`"1234"`)),
		},
		NetworkProfiles: []NetworkProfile{{
			ID: "my-profile-id",
			Data:             json.RawMessage([]byte(`{"consumer_id": "147"}`))}}}
	body, err := json.Marshal(provisionResponse)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(body)).To(MatchJSON(`{
        "abc" : "1234",
		"network_profiles": [{
			"id" : "my-profile-id",
			"data":{
				"consumer_id": "147"
			}
		}]
    }`))
}


func TestProvisionResponseUnmarshalInvalidNetworProfile(t *testing.T) {
	g := NewGomegaWithT(t)
	var provisionResponse ProvisionResponse
	err := json.Unmarshal([]byte(`{
		"network_profiles": 666
    }`), &provisionResponse)
	g.Expect(err).To(HaveOccurred())
}
