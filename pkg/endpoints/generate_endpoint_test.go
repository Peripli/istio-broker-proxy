package endpoints

import (
	"bytes"
	"encoding/json"
	. "github.com/onsi/gomega"
	"testing"
)

func parseResponseData(data []byte, t *testing.T) responseData {
	g := NewGomegaWithT(t)
	var result responseData
	err := json.NewDecoder(bytes.NewReader(data)).Decode(&result)
	g.Expect(err).NotTo(HaveOccurred(),"error while decoding body: %v ", data)

	return result
}

func TestInvalidJSONResultsInError(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{"foo": "bar`)

	_, err := GenerateEndpoint(body)

	g.Expect(err).To(HaveOccurred())
}

func TestUnknownDataIsReturned(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{"foo": "bar"}`)

	newBody, _ := GenerateEndpoint(body)

	g.Expect(newBody).To(Equal(body))
}

func TestEndpointIsCreated(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{
            "credentials": {
               "uri": "postgres://user:pass@mysqlhost:3306/dbname",
               "hostname": "dbhost",
               "port": "3306"
			} }`)
	expected := parseResponseData(body, t)

	newBody, err := GenerateEndpoint(body)
	g.Expect(err).NotTo(HaveOccurred())

	result := parseResponseData(newBody, t)

	g.Expect(result.Endpoints).To(HaveLen(1))
	g.Expect(result.Credentials).To(Equal(expected.Credentials))
}

func TestEndpointDataIsCorrect(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{
            "credentials": {
               "uri": "postgres://user:pass@mysqlhost:3306/dbname",
                "hostname": "dbhost",
                "port": "3306"
			} }`)

	newBody, err := GenerateEndpoint(body)
	g.Expect(err).NotTo(HaveOccurred())

	result := parseResponseData(newBody, t)
	actualPort := result.Endpoints[0]["port"]
	g.Expect(actualPort).To(Equal("3306"))
	actualHost := result.Endpoints[0]["hostname"]
	g.Expect(actualHost).To(Equal("dbhost"))
}

func TestRealWorldExample(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{"credentials":{"hostname":"10.11.241.0","ports":{"5432/tcp":"47637"},"port":"47637","username":"mma4G8N0isoxe17v","password":"tkREVXOktdT2TRF6","dbname":"yLO2WoE0-mCcEppn","uri":"postgres://mma4G8N0isoxe17v:tkREVXOktdT2TRF6@10.11.241.0:47637/yLO2WoE0-mCcEppn"}}`)

	newBody, err := GenerateEndpoint(body)
	g.Expect(err).NotTo(HaveOccurred())

	result := parseResponseData(newBody, t)
	g.Expect(result.Endpoints).To(HaveLen(1))
	actualPort := result.Endpoints[0]["port"]
	g.Expect(actualPort).To(Equal("47637"))
	actualHost := result.Endpoints[0]["hostname"]
	g.Expect(actualHost).To(Equal("10.11.241.0"))
}

func TestIgnoreBlueprintService(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{"credentials":{"hostname":"10.11.241.0","ports":{"8080/tcp":"47818"},"port":"47818","username":"Oqkg-yyjb5Hv_0jJ","password":"vPGsyx0RjNWQ06dF","uri":"http://Oqkg-yyjb5Hv_0jJ:vPGsyx0RjNWQ06dF@10.11.241.0:47818"}}`)

	newBody, err := GenerateEndpoint(body)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(newBody).To(Equal(body))
}
