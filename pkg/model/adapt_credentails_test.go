package model

import (
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
)

const (
	examplePostgresRequest = `{
    "credentials": {
 "dbname": "yLO2WoE0-mCcEppn",
 "hostname": "10.11.241.0",
 "password": "redacted",
 "port": "47637",
 "ports": {"5432/tcp": "47637"},
 "uri": "postgres://mma4G8N0isoxe17v:redacted@10.11.241.0:47637/yLO2WoE0-mCcEppn",
 "username": "mma4G8N0isoxe17v"
},
    "endpoint_mappings": [{
        "source": {"host": "10.11.241.0", "port": 47637},
        "target": {"host": "appnethost", "port": 9876}
	}]
}`
	exampleRabbitMqRequest = `{
    "credentials": {
 "hostname": "10.11.241.0",
 "password": "ypAT7hlpCrvsvzI2",
 "port": "51011",
 "ports": {
  "15672/tcp": "43795",
  "15674/tcp": "39776",
  "15675/tcp": "35827",
  "1883/tcp": "35982",
  "5672/tcp": "51011",
  "61613/tcp": "40865"
 },
 "uri": "amqp://gXoS1dkYUt9lyvZc:ypAT7hlpCrvsvzI2@10.11.241.0:51011",
 "username": "gXoS1dkYUt9lyvZc"
},
    "endpoint_mappings": [{
        "source": {"host": "10.11.241.0", "port": 51011},
        "target": {"host": "appnethost", "port": 9876}
	}]
}`
)

func TestPostgresExampleRequestFromBacklogItem(t *testing.T) {
	g := NewGomegaWithT(t)

	var example AdaptCredentialsRequest
	var expected Credentials
	err := json.Unmarshal([]byte(examplePostgresRequest), &example)
	g.Expect(err).NotTo(HaveOccurred())
	err = json.Unmarshal([]byte(`{
	 "dbname": "yLO2WoE0-mCcEppn",
	 "hostname": "appnethost",
	 "password": "redacted",
	 "port": 9876,
	 "ports": {"5432/tcp": "47637"},
	 "uri": "postgres://mma4G8N0isoxe17v:redacted@appnethost:9876/yLO2WoE0-mCcEppn",
	 "username": "mma4G8N0isoxe17v"
	  }`), &expected)
	g.Expect(err).NotTo(HaveOccurred())

	adapted, _ := Adapt(example.Credentials, example.EndpointMappings)
	g.Expect(adapted.Credentials).To(Equal(expected))
	g.Expect(adapted.Endpoints).To(Equal([]Endpoint{{"appnethost", 9876}}))
}

func TestRabbitMqExampleRequestFromBacklogItem(t *testing.T) {
	g := NewGomegaWithT(t)

	var example AdaptCredentialsRequest
	var expected Credentials
	err := json.Unmarshal([]byte(exampleRabbitMqRequest), &example)
	g.Expect(err).NotTo(HaveOccurred())
	err = json.Unmarshal([]byte(`{
 "hostname": "appnethost",
 "password": "ypAT7hlpCrvsvzI2",
 "port": 9876,
 "ports": {
  "15672/tcp": "43795",
  "15674/tcp": "39776",
  "15675/tcp": "35827",
  "1883/tcp": "35982",
  "5672/tcp": "51011",
  "61613/tcp": "40865"
 },
 "uri": "amqp://gXoS1dkYUt9lyvZc:ypAT7hlpCrvsvzI2@appnethost:9876",
 "username": "gXoS1dkYUt9lyvZc"
}`), &expected)
	g.Expect(err).NotTo(HaveOccurred())

	fmt.Printf("%#v\n", example)
	adapted, _ := Adapt(example.Credentials, example.EndpointMappings)
	g.Expect(adapted.Credentials).To(Equal(expected))
	g.Expect(adapted.Endpoints).To(Equal([]Endpoint{{"appnethost", 9876}}))
}

func TestUnknownCredentials(t *testing.T) {
	g := NewGomegaWithT(t)
	credentials := Credentials{}
	adapted, _ := Adapt(credentials, []EndpointMapping{})
	g.Expect(adapted.Credentials).To(Equal(credentials))
}

func TestEndpointIsAddedAfterApplying(t *testing.T) {
	g := NewGomegaWithT(t)
	var example AdaptCredentialsRequest
	json.Unmarshal([]byte(examplePostgresRequest), &example)

	translatedRequest, _ := Adapt(example.Credentials, example.EndpointMappings)

	g.Expect(len(example.Credentials.Endpoints)).To(Equal(0))
	g.Expect(len(translatedRequest.Endpoints)).To(Equal(1))
	g.Expect(translatedRequest.Endpoints[0]).To(Equal(Endpoint{"appnethost", 9876}))
}
