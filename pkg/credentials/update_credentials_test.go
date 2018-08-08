package credentials

import (
	"encoding/json"
	. "github.com/onsi/gomega"
	"testing"
)

const (
	exampleRequest = `{
    "credentials": {
 "dbname": "yLO2WoE0-mCcEppn",
 "hostname": "10.11.241.0",
 "password": "<redacted>",
 "port": "47637",
 "ports": {
  "5432/tcp": "47637"
 },
 "uri": "postgres://mma4G8N0isoxe17v:<redacted>@10.11.241.0:47637/yLO2WoE0-mCcEppn",
 "username": "mma4G8N0isoxe17v"
},
    "endpoint_mappings": [{
    	"source": {"host": "mysqlhost", "port": 3306},
        "target": {"host": "appnethost", "port": 9876}
	}]
}`
	exampleRequestHaPostgres = `{
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
        "source": {"host": "mysqlhost", "port": 3306},
        "target": {"host": "appnethost", "port": 9876}
	}]
}`
	minimalValidEndpointMapping  = `{ "source":{"host":"a", "port":1}, "target":{"host":"b", "port":2}}`
	minimalValidEndpointMappings = `[` + minimalValidEndpointMapping + `]`
	minimalValidCredentials      = `{ "hostname": "c",  "port": "1", "uri": "postgres://a:b@c:1/d"}`
)

func TestRejectEmptyJson(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(IsValidUpdateRequestBody("")).To(BeFalse())
}

func TestAcceptExampleRequestFromBacklogItem(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(IsValidUpdateRequestBody(exampleRequest)).To(BeTrue())
}

func TestAcceptExampleHaPostgresRequestFromBacklogItem(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(IsValidUpdateRequestBody(exampleRequestHaPostgres)).To(BeTrue())
}

func TestRejectInvalidJson(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(IsValidUpdateRequestBody("{")).To(BeFalse())
}

func TestRejectRequestWithoutCredentials(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(IsValidUpdateRequestBody(`{ "endpoint_mappings": ` + minimalValidEndpointMappings + `}`)).To(BeFalse())
}

func TestRejectRequestWithEmptyCredentials(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(IsValidUpdateRequestBody(`{
    "credentials": {},
    "endpoint_mappings": ` + minimalValidEndpointMappings + `}`)).To(BeFalse())
}

func TestRejectRequestWithoutEndpointMappings(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(IsValidUpdateRequestBody(`{
    "credentials": ` + minimalValidCredentials + `}`)).To(BeFalse())
}

func TestRejectRequestWithEmptyEndpointMappings(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(IsValidUpdateRequestBody(`{ "credentials": ` + minimalValidCredentials +
		`, "endpoint_mappings": [{}] }`)).To(BeFalse())
}

func TestAcceptCredentialsFromBLI(t *testing.T) {
	g := NewGomegaWithT(t)
	var rawJson map[string]interface{}
	json.Unmarshal([]byte(`{
 "dbname": "yLO2WoE0-mCcEppn",
 "hostname": "10.11.241.0",
 "password": "<redacted>",
 "port": "47637",
 "ports": {
  "5432/tcp": "47637"
 },
 "uri": "postgres://mma4G8N0isoxe17v:<redacted>@10.11.241.0:47637/yLO2WoE0-mCcEppn",
 "username": "mma4G8N0isoxe17v"
}`), &rawJson)

	g.Expect(isValidCredentials(rawJson)).To(BeTrue())
}

func TestAcceptMinimalCredentialsAndRejectCredentialsWithMissingFields(t *testing.T) {
	g := NewGomegaWithT(t)
	var validCredentials map[string]interface{}
	json.Unmarshal([]byte(minimalValidCredentials), &validCredentials)
	g.Expect(isValidCredentials(validCredentials)).To(BeTrue())

	var invalidCredentials map[string]interface{} = make(map[string]interface{})
	for _, fieldName := range []string{"uri", "hostname", "port"} {
		copyAllFieldsButOne(validCredentials, invalidCredentials, fieldName)
		g.Expect(isValidCredentials(invalidCredentials)).To(BeFalse())
	}
}

func copyAllFieldsButOne(from map[string]interface{}, to map[string]interface{}, keyToOmit string) {
	for key, value := range from {
		to[key] = value
	}
	delete(to, keyToOmit)
}

func TestRejectEndpointMappingNoArray(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(IsValidUpdateRequestBody(`{
    "credentials": ` + minimalValidCredentials + `,
    "endpoint_mappings": {
    	"source": {"host": "a", "port": 1},
        "target": {"host": "b", "port": 2}
	}
}`)).To(BeFalse())
}

func TestAcceptEmptyEndpointMapping(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(IsValidUpdateRequestBody(`{
    "credentials": ` + minimalValidCredentials + `,
    "endpoint_mappings": []
}`)).To(BeFalse())
}

func TestAcceptMinimalNonEmptyEndpointMappings(t *testing.T) {
	g := NewGomegaWithT(t)
	var validEndpointMappings []interface{}
	json.Unmarshal([]byte(minimalValidEndpointMappings), &validEndpointMappings)
	g.Expect(isValidEndpointMappings(validEndpointMappings)).To(BeTrue())
}

func TestRejectIncompleteEndpointMapping(t *testing.T) {
	g := NewGomegaWithT(t)
	var validEndpointMapping map[string]interface{}
	json.Unmarshal([]byte(minimalValidEndpointMapping), &validEndpointMapping)
	g.Expect(isValidEndpointMapping(validEndpointMapping)).To(BeTrue())

	var invalidEndpointMapping map[string]interface{} = make(map[string]interface{})

	for _, missingField := range []string{"source", "target"} {
		copyAllFieldsButOne(validEndpointMapping, invalidEndpointMapping, missingField)
		g.Expect(isValidEndpointMapping(invalidEndpointMapping)).To(BeFalse())
	}
}

func TestRejectIncompleteEndpoint(t *testing.T) {
	g := NewGomegaWithT(t)
	var invalidEndpointMapping map[string]interface{}
	json.Unmarshal([]byte(`{ "source":{"host":"a"}, "target":{"host":"b", "port":2}}`), &invalidEndpointMapping)
	g.Expect(isValidEndpointMapping(invalidEndpointMapping)).To(BeFalse())
	json.Unmarshal([]byte(`{ "source":{"port":1}, "target":{"host":"b", "port":2}}`), &invalidEndpointMapping)
	g.Expect(isValidEndpointMapping(invalidEndpointMapping)).To(BeFalse())
	json.Unmarshal([]byte(`{ "target":{"host":"a"}, "source":{"host":"b", "port":2}}`), &invalidEndpointMapping)
	g.Expect(isValidEndpointMapping(invalidEndpointMapping)).To(BeFalse())
	json.Unmarshal([]byte(`{ "target":{"port":1}, "source":{"host":"b", "port":2}}`), &invalidEndpointMapping)
	g.Expect(isValidEndpointMapping(invalidEndpointMapping)).To(BeFalse())
}
