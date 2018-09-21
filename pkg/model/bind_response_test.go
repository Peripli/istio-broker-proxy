package model

import (
	"encoding/json"
	. "github.com/onsi/gomega"
	"testing"
)

func TestBindResponseUnmarshal(t *testing.T) {
	g := NewGomegaWithT(t)
	var bindResponse BindResponse
	err := json.Unmarshal([]byte(`{
       "abc" : "1234",
		"network_data": {
			"network_profile_id" : "your-profile-id",
			"data":{
				"provider_id": "852",
				"endpoints":[
				{"host":"10.0.0.1", "port": 9999},
				{"host":"10.0.0.2", "port": 7777}
				]
			}
		},
		"credentials": { "user" : "myuser" },
		"endpoints":[
				{"host":"10.0.0.3", "port": 3333},
				{"host":"10.0.0.4", "port": 4444}
		]
    }`), &bindResponse)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(bindResponse.AdditionalProperties["abc"])).To(Equal(`"1234"`))
	g.Expect(bindResponse.NetworkData.NetworkProfileId).To(Equal("your-profile-id"))
	g.Expect(bindResponse.NetworkData.Data.ProviderId).To(Equal("852"))
	g.Expect(bindResponse.NetworkData.Data.Endpoints[0]).To(Equal(Endpoint{"10.0.0.1", 9999}))
	g.Expect(bindResponse.Endpoints[0]).To(Equal(Endpoint{"10.0.0.3", 3333}))
	g.Expect(bindResponse.Endpoints[1]).To(Equal(Endpoint{"10.0.0.4", 4444}))
	g.Expect(string(bindResponse.Credentials.AdditionalProperties["user"])).To(Equal(`"myuser"`))

}

func TestBindResponseUnmarshalInvalidNetworkData(t *testing.T) {
	g := NewGomegaWithT(t)
	var bindResponse BindResponse
	err := json.Unmarshal([]byte(`{
		"network_data": {
			"network_profile_id" : 1234
		}
    }`), &bindResponse)
	g.Expect(err).To(HaveOccurred())
}

func TestBindResponseUnmarshalInvalidCredentials(t *testing.T) {
	g := NewGomegaWithT(t)
	var bindResponse BindResponse
	err := json.Unmarshal([]byte(`{
		"credentials": { "end_points" : 1234 }
    }`), &bindResponse)
	g.Expect(err).To(HaveOccurred())
}

func TestBindResponseUnmarshalInvalidAdditionalProperties(t *testing.T) {
	g := NewGomegaWithT(t)
	var bindResponse BindResponse
	err := json.Unmarshal([]byte(`[]`), &bindResponse)
	g.Expect(err).To(HaveOccurred())
}

func TestBindResponseUnmarshalInvalidEndpoints(t *testing.T) {
	g := NewGomegaWithT(t)
	var bindResponse BindResponse
	err := json.Unmarshal([]byte(`{
		"endpoints": [ 1234 ]
    }`), &bindResponse)
	g.Expect(err).To(HaveOccurred())
}

func TestBindResponseMarshal(t *testing.T) {
	g := NewGomegaWithT(t)
	bindResponse := BindResponse{
		AdditionalProperties: map[string]json.RawMessage{
			"abc": json.RawMessage([]byte(`"1234"`)),
		},
		NetworkData: NetworkDataResponse{
			NetworkProfileId: "your-profile-id",
			Data: DataResponse{
				ProviderId: "852",
				Endpoints:  []Endpoint{{"host1", 9999}, {"host2", 7777}},
			}},
		Credentials: Credentials{AdditionalProperties: map[string]json.RawMessage{"user": json.RawMessage([]byte(`"myuser"`))}},
		Endpoints:   []Endpoint{{"host3", 3333}, {"host4", 4444}},
	}
	body, err := json.Marshal(bindResponse)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(body)).To(MatchJSON(`{
       "abc" : "1234",
		"network_data": {
			"network_profile_id" : "your-profile-id",
			"data":{
				"provider_id": "852",
				"endpoints":[
				{"host":"host1", "port": 9999},
				{"host":"host2", "port": 7777}
				]
			}
		},
		"credentials": { "user" : "myuser" },
		"endpoints":[
				{"host":"host3", "port": 3333},
				{"host":"host4", "port": 4444}
		]
    }`))
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

	var bindResponse BindResponse
	err := json.Unmarshal(body, &bindResponse)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(bindResponse.Credentials.Endpoints).To(HaveLen(1))
	g.Expect(bindResponse.Credentials.Endpoints[0]).To(Equal(Endpoint{"10.11.241.0", 47637}))
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
	var result BindResponse
	err := json.Unmarshal(body, &result)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(result.Credentials.Endpoints).To(HaveLen(3))

	g.Expect(result.Credentials.Endpoints[0]).To(Equal(Endpoint{"10.11.19.245", 5432}))
	g.Expect(result.Credentials.Endpoints[1]).To(Equal(Endpoint{"10.11.19.240", 5432}))
	g.Expect(result.Credentials.Endpoints[2]).To(Equal(Endpoint{"10.11.19.241", 5432}))
}

func TestIgnoreBlueprintService(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{"credentials":{"hostname":"10.11.241.0","ports":{"8080/tcp":"47818"},"port":47818,"username":"Oqkg-yyjb5Hv_0jJ","password":"vPGsyx0RjNWQ06dF","uri":"http://Oqkg-yyjb5Hv_0jJ:vPGsyx0RjNWQ06dF@10.11.241.0:47818"}}`)

	var result BindResponse
	err := json.Unmarshal(body, &result)
	g.Expect(err).NotTo(HaveOccurred())
	newBody, err := json.Marshal(result)

	g.Expect(newBody).To(MatchJSON(body))
}
