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

	var endpoints []interface{}

	for _, endpointMapping := range endpointMappings {
		endpoints = appendMapping(endpoints, endpointMapping)
	}

	topLevelJson["endpoints"] = endpoints

	bytes, _ := json.Marshal(topLevelJson)
	return string(bytes[:])
}

func appendMapping(endpoints []interface{}, endpointMapping interface{}) []interface{} {
	mapping := toStringMap(endpointMapping)
	target := toStringMap(mapping["target"])
	return append(endpoints, target)
}

func toStringMap(untyped interface{}) map[string]interface{} {
	return untyped.(map[string]interface{})
}

func applyMappings(endpointMappings []interface{}, credentials map[string]interface{}) {

	for _, endpointMappingUntyped := range endpointMappings {
		endpointMapping := toStringMap(endpointMappingUntyped)
		if sourceMatchesCredentials(credentials, endpointMapping) {
			target := toStringMap(endpointMapping["target"])
			credentials["hostname"] = target["host"]
			credentials["port"] = target["port"]
		}
		credentials["uri"] = applyOnUri(credentials["uri"].(string), endpointMapping)
	}
}

func sourceMatchesCredentials(credentials map[string]interface{}, endpointMapping map[string]interface{}) bool {
	return fmt.Sprintf("%v:%v", credentials["hostname"], credentials["port"]) == toHostString(endpointMapping, "source")
}

func applyOnUri(uri string, endpointMapping map[string]interface{}) string {
	parsedUrl := parseUri(uri)
	if sourceMatchesUri(parsedUrl, endpointMapping) {
		parsedUrl.Host = toHostString(endpointMapping, "target")
	}
	return fmt.Sprintf("%v", parsedUrl)
}

func parseUri(uri string) *url.URL {
	url, _ := url.Parse(uri)
	return url
}

func sourceMatchesUri(url *url.URL, endpointMapping map[string]interface{}) bool {
	return url.Host == toHostString(endpointMapping, "source")
}

func toHostString(endpointMapping map[string]interface{}, sourceOrTarget string) string {
	endpoint := toStringMap(endpointMapping[sourceOrTarget])
	return fmt.Sprintf("%v:%v", endpoint["host"], endpoint["port"])
}
