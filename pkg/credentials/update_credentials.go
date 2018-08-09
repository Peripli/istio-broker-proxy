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
	delete(topLevelJson, "endpoint_mappings")
	addEndpoint(topLevelJson, endpointMappings[0])


	bytes, _ := json.Marshal(topLevelJson)
	return string(bytes[:])
}

func addEndpoint(result map[string]interface{}, endpointMappingToAdd interface{}) {
	var endpoints []interface{}

	mapping := toStringMap(endpointMappingToAdd)
	target := toStringMap(mapping["target"])
	endpoints = append(endpoints, target)

	result["endpoints"] = endpoints
}

func toStringMap(untyped interface{}) map[string]interface{} {
	return untyped.(map[string]interface{})
}

func applyMappings(endpointMappings []interface{}, credentials map[string]interface{}) {
	endpointMapping := toStringMap(endpointMappings[0])
	target := toStringMap(endpointMapping["target"])
	credentials["hostname"] = target["host"]
	credentials["port"] = target["port"]
	credentials["uri"] = applyOnUri(credentials["uri"].(string), target["host"], target["port"])
}

func applyOnUri(uri string, host interface{}, port interface{}) string {
	parsedUrl, _ := url.Parse(uri)
	parsedUrl.Host = fmt.Sprintf("%v:%v", host, port)
	return fmt.Sprintf("%v", parsedUrl)
}
