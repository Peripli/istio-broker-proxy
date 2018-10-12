package model

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
)

const (
	exampleRequest = `{
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
	ExampleRequestHaPostgres = `{
    "credentials": {
 "dbname": "e2b91324e12361f3eaeb35fe570efe1d",
 "end_points": [
  {
   "host": "10.11.19.245",
   "network_id": "SF",
   "port": 5432
  },
  {
   "host": "10.11.19.240",
   "network_id": "SF",
   "port": 5432
  },
  {
   "host": "10.11.19.241",
   "network_id": "SF",
   "port": 5432
  }
 ],
 "hostname": "10.11.19.245",
 "password": "c00132ea8771e16c8aecc9a7b819f91c",
 "port": "5432",
 "read_url": "jdbc:postgresql://10.11.19.240,10.11.19.241/e2b91324e12361f3eaeb35fe570efe1d?targetServerType=preferSlave\u0026loadBalanceHosts=true",
 "uri": "postgres://0d158137ea834372c7f7f53036b1faf6:c00132ea8771e16c8aecc9a7b819f91c@10.11.19.245:5432/e2b91324e12361f3eaeb35fe570efe1d",
 "username": "0d158137ea834372c7f7f53036b1faf6",
 "write_url": "jdbc:postgresql://10.11.19.240,10.11.19.241/e2b91324e12361f3eaeb35fe570efe1d?targetServerType=master"
  },
    "endpoint_mappings": [{
        "source": {"host": "10.11.19.245", "port": 5432},
        "target": {"host": "appnethost", "port": 9876}
	}]
}`
	minimalValidEndpointMapping  = `{ "source":{"host":"a", "port":1}, "target":{"host":"b", "port":2}}`
	minimalValidEndpointMappings = `[` + minimalValidEndpointMapping + `]`
)

func TestExampleRequestFromBacklogItem(t *testing.T) {
	g := NewGomegaWithT(t)

	var example AdpotCredentialsRequest
	var expected Credentials
	json.Unmarshal([]byte(exampleRequest), &example)
	json.Unmarshal([]byte(`{
	 "dbname": "yLO2WoE0-mCcEppn",
	 "hostname": "appnethost",
	 "password": "redacted",
	 "port": 9876,
	 "ports": {"5432/tcp": "47637"},
	 "uri": "postgres://mma4G8N0isoxe17v:redacted@appnethost:9876/yLO2WoE0-mCcEppn",
	 "username": "mma4G8N0isoxe17v"
	  }`), &expected)

	adopted, _ := Adopt(example)
	g.Expect(adopted.Credentials).To(Equal(expected))
	g.Expect(adopted.Endpoints).To(Equal([]Endpoint{{"appnethost", 9876}}))
}

func TestEndpointIsAddedAfterApplying(t *testing.T) {
	g := NewGomegaWithT(t)
	var example AdpotCredentialsRequest
	json.Unmarshal([]byte(exampleRequest), &example)

	translatedRequest, _ := Adopt(example)

	g.Expect(len(example.Credentials.Endpoints)).To(Equal(0))
	g.Expect(len(translatedRequest.Endpoints)).To(Equal(1))
	g.Expect(translatedRequest.Endpoints[0]).To(Equal(Endpoint{"appnethost", 9876}))
}
