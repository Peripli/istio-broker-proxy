package profiles

import (
	"encoding/json"
	"fmt"
)

func AddIstioNetworkDataToResponse(providerId string, serviceId string, systemDomain string, portNumber int) func([]byte) ([]byte, error) {
	return func(body []byte) ([]byte, error) {
		var fromJson map[string]interface{}
		err := json.Unmarshal(body, &fromJson)
		if err != nil {
			return nil, err
		}

		endpointCount, err := countEndpoints(fromJson)
		if err != nil {
			return nil, err
		}
		if endpointCount == 0 {
			return body, nil
		}
		endpointHosts := createEndpointHostsBasedOnSystemDomainServiceId(serviceId, systemDomain, endpointCount)

		newEndpoints := make([]map[string]interface{}, 0)
		for _, endpointHost := range endpointHosts {

			newEndpoints = append(newEndpoints, map[string]interface{}{
				"host": endpointHost,
				"port": portNumber,
			},
			)
		}

		fromJson["network_data"] = map[string]interface{}{
			"network_profile_id": "urn:com.sap.istio:public",
			"provider_id":        providerId,
			"endpoints":          newEndpoints,
		}

		newBody, err := json.Marshal(fromJson)

		return newBody, err
	}
}

func countEndpoints(fromJson map[string]interface{}) (int, error) {
	untypedEndpoints := fromJson["endpoints"]
	if untypedEndpoints == nil {
		return 0, nil
	}
	var endpoints []interface{}
	switch untypedEndpoints.(type) {
	case []interface{}:
		endpoints = untypedEndpoints.([]interface{})
	default:
		return 0, fmt.Errorf("request contains invalid endpoints '%v'", untypedEndpoints)
	}
	return len(endpoints), nil
}

func createEndpointHostsBasedOnSystemDomainServiceId(serviceId string, systemDomain string, count int) []string {
	var endpointsHosts []string

	for i := 0; i < count; i++ {
		newHost := fmt.Sprintf("%d.%s.%s", i+1, serviceId, systemDomain)
		endpointsHosts = append(endpointsHosts, newHost)
	}
	return endpointsHosts
}
