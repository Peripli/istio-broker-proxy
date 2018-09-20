package model

import (
	"encoding/json"
	. "github.com/onsi/gomega"
	"testing"
)

func TestCredenialsUnmarshal(t *testing.T) {
	g := NewGomegaWithT(t)
	var credentials Credentials
	err := json.Unmarshal([]byte(`{
        "user" : "myuser",
		"end_points":[
				{"host":"10.0.0.3", "port": "3333"},
				{"host":"10.0.0.4", "port": "4444"}
		]
    }`), &credentials)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(credentials.AdditionalProperties["user"])).To(Equal(`"myuser"`))
	g.Expect(credentials.Endpoints[0]).To(Equal(Endpoint{"10.0.0.3", 3333}))
	g.Expect(credentials.Endpoints[1]).To(Equal(Endpoint{"10.0.0.4", 4444}))

}

func TestCredentialsMarshal(t *testing.T) {
	g := NewGomegaWithT(t)
	credentials := Credentials{
		AdditionalProperties: map[string]json.RawMessage{
			"user": json.RawMessage([]byte(`"myuser"`)),
		},
		Endpoints: []Endpoint{{"10.0.0.3", 3333}, {"10.0.0.4", 4444}},
	}
	body, err := json.Marshal(credentials)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(body)).To(MatchJSON(`{
		"end_points":[
				{"host":"10.0.0.3", "port": 3333},
				{"host":"10.0.0.4", "port": 4444}
		],
        "user" : "myuser"
    }`))
}

func TestCredentialsInvalidAdditionalProperties(t *testing.T) {
	g := NewGomegaWithT(t)
	var credentials Credentials
	err := json.Unmarshal([]byte(`[]`), &credentials)
	g.Expect(err).To(HaveOccurred())
}
