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
 "password": "redacted",
 "port": "47637",
 "ports": {
  "5432/tcp": "47637"
 },
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

func TestRejectEmptyJson(t *testing.T) {
	g := NewGomegaWithT(t)
	ok, e := isValidUpdateRequestBody("")
	g.Expect(e).Should(HaveOccurred())
	g.Expect(ok).To(BeFalse())
}

func TestAcceptExampleRequestFromBacklogItem(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(isValidUpdateRequestBody(exampleRequest)).To(BeTrue())
}

func TestAcceptExampleHaPostgresRequestFromBacklogItem(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(isValidUpdateRequestBody(ExampleRequestHaPostgres)).To(BeTrue())
}

func TestRejectInvalidJson(t *testing.T) {
	g := NewGomegaWithT(t)
	ok, e := isValidUpdateRequestBody("{")
	g.Expect(e).Should(HaveOccurred())
	g.Expect(ok).To(BeFalse())
}

func TestRejectRequestWithoutCredentials(t *testing.T) {
	g := NewGomegaWithT(t)
	ok, e := isValidUpdateRequestBody(`{ "endpoint_mappings": ` + minimalValidEndpointMappings + `}`)
	g.Expect(e).Should(HaveOccurred())
	g.Expect(ok).To(BeFalse())
}

func TestRejectRequestWithEmptyCredentials(t *testing.T) {
	g := NewGomegaWithT(t)
	ok, e := isValidUpdateRequestBody(`{
	    "credentials": {},
	    "endpoint_mappings": ` + minimalValidEndpointMappings + `}`)
	g.Expect(e).Should(HaveOccurred())
	g.Expect(ok).To(BeFalse())
}

func TestRejectRequestWithoutEndpointMappings(t *testing.T) {
	g := NewGomegaWithT(t)
	_, e := isValidUpdateRequestBody(`{
	    "credentials": ` + (`{ "hostname": "c",  "port": "1", "uri": "postgres://a:b@c:1/d"}`) + `}`)
	g.Expect(e).Should(HaveOccurred())
}

func TestRejectRequestWithEmptyEndpointMappings(t *testing.T) {
	g := NewGomegaWithT(t)
	ok, e := isValidUpdateRequestBody(`{ "credentials": ` + (`{ "hostname": "c",  "port": "1", "uri": "postgres://a:b@c:1/d"}`) +
		`, "endpoint_mappings": [{}] }`)
	g.Expect(e).Should(HaveOccurred())
	g.Expect(ok).To(BeFalse())
}

func TestRejectRequestWithNonMatchingEndpointMappings(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(isValidUpdateRequestBody(`{ "credentials": ` + (`{ "hostname": "c",  "port": "1", "uri": "postgres://a:b@c:1/d"}`) +
		`, "endpoint_mappings": ` + minimalValidEndpointMappings + ` }`)).To(BeFalse())
}

func TestRejectRequestWithSecondEndpointMappingIsNotMatching(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(isValidUpdateRequestBody(`{ "credentials": ` + (`{ "hostname": "c",  "port": "1", "uri": "postgres://a:b@c:1/d"}`) +
		`, "endpoint_mappings": [ { "source": { "host": "c", "port": 1}, "target": { "host": "d", "port": 2} },` +
		minimalValidEndpointMapping + `]}`)).To(BeFalse())
}

func TestAcceptCredentialsFromBLI(t *testing.T) {
	g := NewGomegaWithT(t)
	var rawJson map[string]interface{}
	json.Unmarshal([]byte(`{
 "dbname": "yLO2WoE0-mCcEppn",
 "hostname": "10.11.241.0",
 "password": "redacted",
 "port": "47637",
 "ports": {
  "5432/tcp": "47637"
 },
 "uri": "postgres://mma4G8N0isoxe17v:redacted@10.11.241.0:47637/yLO2WoE0-mCcEppn",
 "username": "mma4G8N0isoxe17v"
}`), &rawJson)

	g.Expect(isValidCredentials(rawJson)).To(BeTrue())
}

func TestAcceptMinimalCredentialsAndRejectCredentialsWithMissingFields(t *testing.T) {
	g := NewGomegaWithT(t)
	var validCredentials map[string]interface{}
	json.Unmarshal([]byte((`{ "hostname": "c",  "port": "1", "uri": "postgres://a:b@c:1/d"}`)), &validCredentials)
	g.Expect(isValidCredentials(validCredentials)).To(BeTrue())

	var invalidCredentials map[string]interface{} = make(map[string]interface{})
	for _, fieldName := range []string{"uri", "hostname", "port"} {
		copyAllFieldsButOne(validCredentials, invalidCredentials, fieldName)
		ok, e := isValidCredentials(invalidCredentials)
		g.Expect(e).Should(HaveOccurred())
		g.Expect(ok).To(BeFalse())
	}
}

func copyAllFieldsButOne(from map[string]interface{}, to map[string]interface{}, keyToOmit string) {
	for key, value := range from {
		to[key] = value
	}
	delete(to, keyToOmit)
}

func TestRejectInvalidUri(t *testing.T) {
	g := NewGomegaWithT(t)
	var credentials map[string]interface{}
	json.Unmarshal([]byte(`{ "hostname": "c",  "port": "1", "uri": "postgres://a:<b>@c:1/d"}`), &credentials)

	ok, e := isValidCredentials(credentials)
	g.Expect(e).Should(HaveOccurred())
	g.Expect(ok).To(BeFalse())
}

func TestRejectEndpointMappingNoArray(t *testing.T) {
	g := NewGomegaWithT(t)
	ok, e := isValidUpdateRequestBody(`{
	    "credentials": ` + (`{ "hostname": "c",  "port": "1", "uri": "postgres://a:b@c:1/d"}`) + `,
	    "endpoint_mappings": {
	      "source": {"host": "a", "port": 1},
	      "target": {"host": "b", "port": 2}
		}
	}`)
	g.Expect(e).Should(HaveOccurred())
	g.Expect(ok).To(BeFalse())
}

func TestRejectEmptyEndpointMapping(t *testing.T) {
	g := NewGomegaWithT(t)
	ok, e := isValidUpdateRequestBody(`{
	    "credentials": ` + (`{ "hostname": "c",  "port": "1", "uri": "postgres://a:b@c:1/d"}`) + `,
	    "endpoint_mappings": []
	}`)
	g.Expect(e).Should(HaveOccurred())
	g.Expect(ok).To(BeFalse())
}

func TestAcceptMinimalNonEmptyEndpointMappings(t *testing.T) {
	g := NewGomegaWithT(t)
	var validEndpointMappings []interface{}
	json.Unmarshal([]byte(minimalValidEndpointMappings), &validEndpointMappings)
	g.Expect(isValidEndpointMappings(validEndpointMappings)).To(BeTrue())
}

func TestAcceptTwoEndpointMappings(t *testing.T) {
	g := NewGomegaWithT(t)
	var validEndpointMappings []interface{}
	json.Unmarshal([]byte(`[`+minimalValidEndpointMapping+`,`+minimalValidEndpointMapping+`]`), &validEndpointMappings)

	ok, e := isValidEndpointMappings(validEndpointMappings)
	g.Expect(e).ShouldNot(HaveOccurred())
	g.Expect(ok).To(BeTrue())
}

func TestRejectIncompleteEndpointMapping(t *testing.T) {
	g := NewGomegaWithT(t)
	var validEndpointMapping map[string]interface{}
	json.Unmarshal([]byte(minimalValidEndpointMapping), &validEndpointMapping)
	g.Expect(isValidEndpointMapping(validEndpointMapping)).To(BeTrue())

	var invalidEndpointMapping map[string]interface{} = make(map[string]interface{})

	for _, missingField := range []string{"source", "target"} {
		copyAllFieldsButOne(validEndpointMapping, invalidEndpointMapping, missingField)
		ok, e := isValidEndpointMapping(invalidEndpointMapping)
		g.Expect(e).Should(HaveOccurred())
		g.Expect(ok).To(BeFalse())
	}
}

func TestRejectIncompleteEndpoint(t *testing.T) {
	g := NewGomegaWithT(t)
	var invalidEndpointMapping map[string]interface{}

	json.Unmarshal([]byte(`{ "source":{"host":"a"}, "target":{"host":"b", "port":2}}`), &invalidEndpointMapping)
	ok, e := isValidEndpointMapping(invalidEndpointMapping)
	g.Expect(e).Should(HaveOccurred())
	g.Expect(ok).To(BeFalse())

	json.Unmarshal([]byte(`{ "source":{"port":1}, "target":{"host":"b", "port":2}}`), &invalidEndpointMapping)
	ok, e = isValidEndpointMapping(invalidEndpointMapping)
	g.Expect(e).Should(HaveOccurred())
	g.Expect(ok).To(BeFalse())

	json.Unmarshal([]byte(`{ "target":{"host":"a"}, "source":{"host":"b", "port":2}}`), &invalidEndpointMapping)
	ok, e = isValidEndpointMapping(invalidEndpointMapping)
	g.Expect(e).Should(HaveOccurred())
	g.Expect(ok).To(BeFalse())

	json.Unmarshal([]byte(`{ "target":{"port":1}, "source":{"host":"b", "port":2}}`), &invalidEndpointMapping)
	ok, e = isValidEndpointMapping(invalidEndpointMapping)
	g.Expect(e).Should(HaveOccurred())
	g.Expect(ok).To(BeFalse())
}

func TestShouldApplyMatchingEndpoint(t *testing.T) {
	g := NewGomegaWithT(t)

	var validEndpointMapping map[string]interface{}
	json.Unmarshal([]byte(`{"source": {"host":"c", "port":1}}`), &validEndpointMapping)

	var validCredentials map[string]interface{}
	json.Unmarshal([]byte((`{ "hostname": "c",  "port": "1", "uri": "postgres://a:b@c:1/d"}`)), &validCredentials)

	g.Expect(shouldApply(validEndpointMapping, validCredentials)).To(BeTrue())
}

func TestShouldApplyEndpointMatchOnHostnameAndPort(t *testing.T) {
	g := NewGomegaWithT(t)

	var validEndpointMapping map[string]interface{}
	json.Unmarshal([]byte(`{"source":{"host":"c", "port":1}}`), &validEndpointMapping)

	var validCredentials map[string]interface{}
	json.Unmarshal([]byte((`{ "hostname": "c",  "port": "1", "uri": "postgres://a:b@xx:99/d"}`)), &validCredentials)

	g.Expect(shouldApply(validEndpointMapping, validCredentials)).To(BeTrue())
}

func TestShouldApplyEndpointMatchOnUri(t *testing.T) {
	g := NewGomegaWithT(t)

	var validEndpointMapping map[string]interface{}
	json.Unmarshal([]byte(`{"source":{"host":"c", "port":1}}`), &validEndpointMapping)

	var validCredentials map[string]interface{}
	json.Unmarshal([]byte((`{ "hostname": "xx",  "port": "99", "uri": "postgres://a:b@c:1/d"}`)), &validCredentials)

	g.Expect(shouldApply(validEndpointMapping, validCredentials)).To(BeTrue())
}
