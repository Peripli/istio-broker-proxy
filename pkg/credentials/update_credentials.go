package credentials

import (
	"encoding/json"
	"fmt"
	"net/url"
)

func translateCredentials(request string) string {
	var topLevelJson map[string]interface{}
	json.Unmarshal([]byte(request), &topLevelJson)
	credentials := toStringMap(topLevelJson["credentials"])
	endpointMappings := topLevelJson["endpoint_mappings"].([]interface{})
	applyMappings(endpointMappings, credentials)
	bytes, _ := json.Marshal(topLevelJson)
	return string(bytes[:])
}

func toStringMap(untyped interface{}) map[string]interface{} {
	return untyped.(map[string]interface{})
}

func applyMappings(endpointMappings []interface{}, credentials map[string]interface{}) {
	for _, endpointMappingUntyped := range endpointMappings {
		endpointMapping := toStringMap(endpointMappingUntyped)
		if shouldApply(toStringMap(endpointMapping["source"]), credentials) {
			target := toStringMap(endpointMapping["target"])
			credentials["hostname"] = target["host"]
			credentials["port"] = target["port"]
			credentials["uri"] = applyOnUri(credentials["uri"].(string), target["host"], target["port"])
			break
		}
	}
}

func applyOnUri(uri string, host interface{}, port interface{}) string {
	parsedUrl, _ := url.Parse(uri)
	parsedUrl.Host = fmt.Sprintf("%v:%v", host, port)
	return fmt.Sprintf("%v", parsedUrl)
}

func shouldApply(source map[string]interface{}, credentials map[string]interface{}) bool {
	sourcePort := fmt.Sprintf("%v", source["port"])
	credentialsPort := fmt.Sprintf("%v", credentials["port"])
	return source["host"] == credentials["hostname"] && sourcePort == credentialsPort
}
