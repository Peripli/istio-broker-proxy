package profiles

import (
	"encoding/json"
	"errors"
	"fmt"
)

func AddIstioNetworkDataToResponse(providerId string, serviceId string, systemDomain string, portNumber int) func([]byte) ([]byte, error) {
	return func(body []byte) ([]byte, error) {
		var fromJson map[string]interface{}
		err := json.Unmarshal(body, &fromJson)
		if err != nil {
			return nil, err
		}

		if fromJson["endpoints"] == nil {
			return body, nil
		}

		endpoints := fromJson["endpoints"].([]interface{})
		endpointHosts, err := createEndpointHostsBasedOnSystemDomainServiceId(serviceId, systemDomain, endpoints)

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

func createEndpointHostsBasedOnSystemDomainServiceId(serviceId string, systemDomain string, endpoints []interface{}) ([]string, error) {
	var endpointsHost []string

	if endpoints == nil {
		return endpointsHost, errors.New("no valid endpoints given!")
	}

	for i, _ := range endpoints {
		epIndex := i + 1
		newHost := fmt.Sprintf("%d.%s.%s", epIndex, serviceId, systemDomain)
		endpointsHost = append(endpointsHost, newHost)
	}
	return endpointsHost, nil
}
