package profiles

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
)

func AddIstioNetworkDataToResponse(providerId string, ctx *gin.Context, systemDomain string) func([]byte) ([]byte, error) {

	serviceId := ctx.Params.ByName("instance_id")

	return func(body []byte) ([]byte, error) {
		var fromJson map[string]interface{}
		err := json.Unmarshal(body, &fromJson)
		if err != nil {
			return nil, err
		}

		//ep := fromJson["credentials"].(map[string]interface{})
		//endpoints := []string{fmt.Sprintf("%v", ep["port"])}
		//endpointHost, err := createEndpointHostsBasedOnSystemDomainServiceId(serviceId, systemDomain, endpoints)

		ep := fromJson["endpoints"].([]interface{})
		endpointHost, err := createEndpointHostsBasedOnSystemDomainServiceId(serviceId, systemDomain, ep)

		newEndpoints := make([]map[string]interface{}, 0)
		for _, eph := range endpointHost {

			newEndpoints = append(newEndpoints, map[string]interface{}{
				"host": eph,
				"port": 9000, //ToDo aus endpoints auslesen
			},
			)
		}

		fromJson["network_data"] = map[string]interface{}{
			"network_profile_id": "urn:com.sap.istio:public",
			"provider_id":        providerId,
			"endpoints":          newEndpoints,
		}

		//fmt.Printf("%v\n", fromJson)

		newBody, err := json.Marshal(fromJson)

		return newBody, err
	}
}

func CreateConfigurableNetworkProfile(providerId string, serviceId string, systemDomain string, endpoints []interface{}) func([]byte) ([]byte, error) {

	//serviceId := ctx.Params.ByName*"instance_id"
	endpointHost, err := createEndpointHostsBasedOnSystemDomainServiceId(serviceId, systemDomain, endpoints)

	if nil != err { //Todo remove when adding data works and adjust test that has only one parameter
		return func(body []byte) ([]byte, error) {
			var fromJson map[string]interface{}
			err := json.Unmarshal(body, &fromJson)
			if err != nil {
				return nil, err
			}

			fromJson["network_data"] = map[string]interface{}{
				"network_profile_id": "urn:com.sap.istio:public",
				"provider_id":        providerId,
			}

			newBody, err := json.Marshal(fromJson)

			return newBody, err
		}
	}

	return func(body []byte) ([]byte, error) {
		var fromJson map[string]interface{}
		err := json.Unmarshal(body, &fromJson)
		if err != nil {
			return nil, err
		}

		newEndpoints := make([]map[string]interface{}, 0)
		for _, eph := range endpointHost {

			newEndpoints = append(newEndpoints, map[string]interface{}{
				"host": eph,
				"port": 9000, //ToDo aus endpoints auslesen
			},
			)
		}

		fromJson["network_data"] = map[string]interface{}{
			"network_profile_id": "urn:com.sap.istio:public",
			"provider_id":        providerId,
			"endpoints":          newEndpoints,
		}

		fmt.Printf("%v\n", fromJson)

		newBody, err := json.Marshal(fromJson)

		return newBody, err
	}
}

func AddIstioNetworkData(body []byte) ([]byte, error) {
	var fromJson map[string]interface{}
	err := json.Unmarshal(body, &fromJson)
	if err != nil {
		return nil, err
	}

	fromJson["network_data"] = map[string]interface{}{
		"network_profile_id": "urn:com.sap.istio:public",
		"provider_id":        "my-provider",
	}

	newBody, err := json.Marshal(fromJson)

	return newBody, err
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
