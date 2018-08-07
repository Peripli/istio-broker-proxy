package credentials

import (
	"encoding/json"
	. "github.com/onsi/gomega"
	"testing"
		)

func TestRejectEmptyJson(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(IsValidUpdateRequestBody("")).To(BeFalse())
}

func TestAcceptExampleRequestFromBacklogItem(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(IsValidUpdateRequestBody(`{
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
}`)).To(BeTrue())
}

func TestRejectInvalidJson(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(IsValidUpdateRequestBody("{")).To(BeFalse())
}

func TestRejectRequestWithoutCredentials(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(IsValidUpdateRequestBody(`{
   "endpoint_mappings": [{
   	"source": {"host": "mysqlhost", "port": 3306},
       "target": {"host": "appnethost", "port": 9876}
	}]
}`)).To(BeFalse())
}

func TestRejectRequestWithoutEndpointMappings(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(IsValidUpdateRequestBody(`{
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
}
}`)).To(BeFalse())
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
	json.Unmarshal([]byte(`{ "hostname": "c",  "port": "1", "uri": "postgres://a:b@c:1/d"}`), &validCredentials)
	g.Expect(isValidCredentials(validCredentials)).To(BeTrue())

	var invalidCredentials map[string]interface{} = make(map[string]interface{})
	for _, fieldName := range []string{ "uri", "hostname", "port"} {
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
    "credentials":{ "hostname": "c",  "port": "1", "uri": "postgres://a:b@c:1/d" },
    "endpoint_mappings": {
    	"source": {"host": "a", "port": 1},
        "target": {"host": "b", "port": 2}
	}
}`)).To(BeFalse())
}

func TestAcceptEmptyEndpointMapping(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(IsValidUpdateRequestBody(`{
    "credentials":{ "hostname": "c",  "port": "1", "uri": "postgres://a:b@c:1/d" },
    "endpoint_mappings": []
}`)).To(BeFalse())
}

func TestAcceptMinimalNonEmptyEndpointMappings(t *testing.T) {
	g := NewGomegaWithT(t)
	var validEndpointMappings []interface{}
	json.Unmarshal([]byte(`[{ "source":{"host":"a", "port":1}, "target":{"host":"b", "port":2}}]`), &validEndpointMappings)
	g.Expect(isValidEndpointMappings(validEndpointMappings)).To(BeTrue())
}

func TestRejectIncompleteEndpointMapping(t *testing.T) {
	g := NewGomegaWithT(t)
	var validEndpointMapping map[string]interface{}
	json.Unmarshal([]byte(`{ "source":{"host":"a", "port":1}, "target":{"host":"b", "port":2}}`), &validEndpointMapping)
	g.Expect(isValidEndpointMapping(validEndpointMapping)).To(BeTrue())

	var invalidEndpointMapping map[string]interface{} = make(map[string]interface{})

	for _, missingField := range []string { "source", "target" } {
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
