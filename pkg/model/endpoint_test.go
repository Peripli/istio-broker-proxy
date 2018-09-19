package model

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
	g.Expect(err).NotTo(HaveOccurred(), "error while decoding body: %v ", data)

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
               "uri": "postgres://user:pass@dbhost:3306/dbname",
               "hostname": "10.1.0.1",
                "port": "3306",
                "end_points": [
                    {
                        "host": "10.1.0.1",
                        "network_id": "SF",
                        "port": 3306
                    }
                ]
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
               "uri": "postgres://user:pass@dbhost:3306/dbname",
                "hostname": "0.1.0.1",
                "port": "3306",
                "end_points": [
                    {
                        "host": "0.1.0.1",
                        "network_id": "SF",
                        "port": 3306
                    }
                ]
            } }`)

	newBody, err := GenerateEndpoint(body)
	g.Expect(err).NotTo(HaveOccurred())

	result := parseResponseData(newBody, t)
	g.Expect(result.Endpoints).To(HaveLen(1))
	g.Expect(result.Endpoints[0]).To(Equal(Endpoint{"0.1.0.1", 3306}))
}

func TestRealWorldExample(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{
    "credentials": {
        "hostname": "10.11.241.0",
        "ports": {
            "5432/tcp": "47637"
        },
        "end_points": [
            {
                "host": "10.11.241.0",
                "network_id": "SF",
                "port": 47637
            }
        ],
        "port": "47637",
        "username": "mma4G8N0isoxe17v",
        "password": "tkREVXOktdT2TRF6",
        "dbname": "yLO2WoE0-mCcEppn",
        "uri": "postgres://mma4G8N0isoxe17v:tkREVXOktdT2TRF6@10.11.241.0:47637/yLO2WoE0-mCcEppn"
    }
    }`)

	newBody, err := GenerateEndpoint(body)
	g.Expect(err).NotTo(HaveOccurred())

	result := parseResponseData(newBody, t)
	g.Expect(result.Endpoints).To(HaveLen(1))
	g.Expect(result.Endpoints[0]).To(Equal(Endpoint{"10.11.241.0", 47637}))
}

func TestInvalidEndpoints(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{
    "credentials": {
        "end_points": ["invalid"],
        "uri": "postgres://mma4G8N0isoxe17v:tkREVXOktdT2TRF6@10.11.241.0:47637/yLO2WoE0-mCcEppn"
    }
    }`)

	_, err := GenerateEndpoint(body)
	g.Expect(err).To(HaveOccurred())
}

func TestHaPostgresExample(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{
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
    }
    }`)
	newBody, err := GenerateEndpoint(body)
	g.Expect(err).NotTo(HaveOccurred())

	result := parseResponseData(newBody, t)
	g.Expect(result.Endpoints).To(HaveLen(3))

	g.Expect(result.Endpoints[0]).To(Equal(Endpoint{"10.11.19.245", 5432}))
	g.Expect(result.Endpoints[1]).To(Equal(Endpoint{"10.11.19.240", 5432}))
	g.Expect(result.Endpoints[2]).To(Equal(Endpoint{"10.11.19.241", 5432}))
}

func TestIgnoreBlueprintService(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{"credentials":{"hostname":"10.11.241.0","ports":{"8080/tcp":"47818"},"port":"47818","username":"Oqkg-yyjb5Hv_0jJ","password":"vPGsyx0RjNWQ06dF","uri":"http://Oqkg-yyjb5Hv_0jJ:vPGsyx0RjNWQ06dF@10.11.241.0:47818"}}`)

	newBody, err := GenerateEndpoint(body)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(newBody).To(Equal(body))
}

func TestEndpointUnmarshalPortAsString(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{
                "host": "10.11.19.245",
                "network_id": "SF",
                "port": "5432"
            }`)
	var ep Endpoint
	err := json.Unmarshal(body, &ep)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ep.Port).To(Equal(5432))
}

func TestEndpointUnmarshalPortAsIntg(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{
                "host": "10.11.19.245",
                "network_id": "SF",
                "port": 5432
            }`)
	var ep Endpoint
	err := json.Unmarshal(body, &ep)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ep.Port).To(Equal(5432))
}

func TestEndpointHostIsIP(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{
                "host": "my-host",
                "network_id": "SF",
                "port": "5432"
            }`)
	var ep Endpoint
	err := json.Unmarshal(body, &ep)
	g.Expect(err).ToNot(HaveOccurred())
}
