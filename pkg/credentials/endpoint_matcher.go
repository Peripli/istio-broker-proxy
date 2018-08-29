package credentials

import (
	"encoding/json"
	"fmt"
	"github.com/onsi/gomega/types"
)

func hasEndpointMappings(jsonString string) bool {
	var fromJson map[string]interface{}
	err := json.Unmarshal([]byte(jsonString), &fromJson)
	if err != nil {
		return false
	}

	mappingFound := (fromJson["endpoint_mappings"] != nil)
	return mappingFound
}

func haveTheEndpoint(host string, port string) types.GomegaMatcher {
	return EndpointMatcher{host, port}
}

type EndpointMatcher struct {
	expectedHost string
	expectedPort string
}

func (e EndpointMatcher) Match(actual interface{}) (returnValue bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			returnValue = false
			err = nil
		}
	}()

	text := actual.(string)
	var fromJson map[string]interface{}
	err = json.Unmarshal([]byte(text), &fromJson)
	if err != nil {
		return false, err
	}

	endpointsUntyped := fromJson["endpoints"]
	if endpointsUntyped == nil {
		return false, nil
	}
	endpoints := endpointsUntyped.([]interface{})

	for _, endpointUntyped := range endpoints {
		if e.isCorrectEndpoint(endpointUntyped) {
			return true, nil
		}
	}

	return false, nil
}

func (e EndpointMatcher) isCorrectEndpoint(endpointUntyped interface{}) bool {
	endpoint := toStringMap(endpointUntyped)
	portAsString := asString(endpoint["port"])
	match := (endpoint["host"].(string) == e.expectedHost) && (portAsString == e.expectedPort)
	return match
}

func asString(in interface{}) string {
	return fmt.Sprintf("%v", in)
}

func (e EndpointMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Endpoint not found (host: %s, port: %s)\nActual: %v", e.expectedHost, e.expectedPort, actual)
}

func (e EndpointMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Endpoint found, but not expected (host: %s, port: %s)\nActual: %v", e.expectedHost, e.expectedPort, actual)
}
