package credentials

import (
	"encoding/json"
	"fmt"
	"os"
	"net/url"
)

func isValidUpdateRequestBody(request string) bool {
	var rawJson interface{}
	err := json.Unmarshal([]byte(request), &rawJson)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid request \"%s\", unmarshalling results in error: %v\n", request, err)
		return false
	}
	topLevelJson := rawJson.(map[string]interface{})
	if !hasField(topLevelJson, "credentials") {
		return false
	}
	if !hasField(topLevelJson, "endpoint_mappings") {
		return false
	}
	credentials := topLevelJson["credentials"].(map[string]interface{})
	if !isValidCredentials(credentials) {
		return false
	}
	endpointMappings := topLevelJson["endpoint_mappings"]
	if !isValidEndpointMappings(endpointMappings) {
		return false
	}
	if !shouldApply(toStringMap(toStringMap(endpointMappings.([]interface{})[0])["source"]), credentials) {
		return false
	}
	return true
}

func isValidEndpointMappings(endpointMappings interface{}) bool {
	switch endpointMappings.(type) {
	case []interface{}:
		if len(endpointMappings.([]interface{})) != 1 {
			fmt.Fprintf(os.Stderr, "Field \"%s\", must contain exactly one mapping.\n", "endpoint_mappings")
			return false
		}
		for _, value := range endpointMappings.([]interface{}) {
			if !isValidEndpointMapping(value.(map[string]interface{})) {
				return false
			}
		}
		return true
	default:
		fmt.Fprintf(os.Stderr, "Invalid type of \"%v\" as field \"%s\", it must be an array.\n", endpointMappings, "endpoint_mappings")
		return false
	}
}

func isValidEndpointMapping(endpointMapping map[string]interface{}) bool {
	if !hasField(endpointMapping, "source") {
		return false
	}
	if !hasField(endpointMapping, "target") {
		return false
	}
	if !isValidEndpoint(endpointMapping["source"].(map[string]interface{})) {
		return false
	}
	if !isValidEndpoint(endpointMapping["target"].(map[string]interface{})) {
		return false
	}
	return true
}

func isValidEndpoint(jsonMap map[string]interface{}) bool {
	return hasField(jsonMap, "host") && hasField(jsonMap, "port")
}

func isValidCredentials(jsonMap map[string]interface{}) bool {
	return hasField(jsonMap, "uri") &&
		hasField(jsonMap, "hostname") &&
		hasField(jsonMap, "port") &&
		canParseUri(jsonMap["uri"].(string))
}

func canParseUri(uriValue string) bool {
	_, err := url.Parse(uriValue)
	return err == nil
}

func hasField(jsonMap map[string]interface{}, fieldName string) bool {
	if jsonMap[fieldName] == nil {
		fmt.Fprintf(os.Stderr, "Invalid json map \"%s\", it contains no \"%s\"\n", jsonMap, fieldName)
		return false
	}
	return true
}

func shouldApply(source map[string]interface{}, credentials map[string]interface{}) bool {
	sourcePort := fmt.Sprintf("%v", source["port"])
	credentialsPort := fmt.Sprintf("%v", credentials["port"])
	return source["host"] == credentials["hostname"] && sourcePort == credentialsPort
}
