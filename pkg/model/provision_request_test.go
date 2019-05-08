package model

import (
	"encoding/json"
	. "github.com/onsi/gomega"
	"testing"
)

func TestProvisionRequestUnmarshal(t *testing.T) {
	g := NewGomegaWithT(t)
	var provisionRequest ProvisionRequest
	err := json.Unmarshal([]byte(`{
       "abc" : "1234",
		"network_profiles": [{
			"id" : "my-profile-id",
			"data":{
				"consumer_id": "147"
			}
		}]
    }`), &provisionRequest)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(provisionRequest.AdditionalProperties["abc"])).To(Equal(`"1234"`))
	g.Expect(string(provisionRequest.NetworkProfiles[0].ID)).To(Equal("my-profile-id"))
	g.Expect(string(provisionRequest.NetworkProfiles[0].Data)).To(MatchJSON(`{"consumer_id": "147"}`))
}

func TestProvisionRequestMarshal(t *testing.T) {
	g := NewGomegaWithT(t)
	var provisionRequest = ProvisionRequest{
		AdditionalProperties: map[string]json.RawMessage{
			"abc": json.RawMessage([]byte(`"1234"`)),
		},
		NetworkProfiles: []NetworkProfile{{
			ID: "my-profile-id",
			Data:             json.RawMessage([]byte(`{"consumer_id": "147"}`))}}}
	body, err := json.Marshal(provisionRequest)
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


func TestProvisionRequestUnmarshalInvalidNetworkProfile(t *testing.T) {
	g := NewGomegaWithT(t)
	var provisionRequest ProvisionRequest
	err := json.Unmarshal([]byte(`{
		"network_profiles": 666
    }`), &provisionRequest)
	g.Expect(err).To(HaveOccurred())
}
