package credentials

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

const default_postgres_port = "5432"
const key_endpoint_mapping = "endpoint_mappings"
const key_credentials = "credentials"
const key_target = "target"
const key_hostname = "hostname"
const key_host = "host"
const key_port = "port"
const key_source = "source"
const key_write_url = "write_url"
const key_read_url = "read_url"
const key_uri = "uri"
const key_endpoints = "endpoints"

func translateCredentials(request string) string {
	var topLevelJson map[string]interface{}
	json.Unmarshal([]byte(request), &topLevelJson)
	credentials := toStringMap(topLevelJson[key_credentials])

	endpointMappings := topLevelJson[key_endpoint_mapping].([]interface{})

	applyMappings(endpointMappings, credentials)
	delete(topLevelJson, key_endpoint_mapping)

	var endpoints []interface{}

	for _, endpointMapping := range endpointMappings {
		endpoints = appendMapping(endpoints, endpointMapping)
	}

	topLevelJson[key_endpoints] = endpoints

	bytes, _ := json.Marshal(topLevelJson)
	return string(bytes[:])
}

func appendMapping(endpoints []interface{}, endpointMapping interface{}) []interface{} {
	mapping := toStringMap(endpointMapping)
	target := toStringMap(mapping[key_target])
	return append(endpoints, target)
}

func toStringMap(untyped interface{}) map[string]interface{} {
	return untyped.(map[string]interface{})
}

func applyMappings(endpointMappings []interface{}, credentials map[string]interface{}) {
	for _, endpointMappingUntyped := range endpointMappings {
		endpointMapping := toStringMap(endpointMappingUntyped)
		if sourceMatchesCredentials(credentials, endpointMapping) {
			target := toStringMap(endpointMapping[key_target])
			credentials[key_hostname] = target[key_host]
			credentials[key_port] = target[key_port]
		}
		credentials[key_uri] = applyOnUri(credentials[key_uri].(string), endpointMapping)
		applyWriteReadUrlIfExists(credentials, endpointMapping, key_write_url)
		applyWriteReadUrlIfExists(credentials, endpointMapping, key_read_url)
	}
}

func sourceMatchesCredentials(credentials map[string]interface{}, endpointMapping map[string]interface{}) bool {
	return fmt.Sprintf("%v:%v", credentials[key_hostname], credentials[key_port]) == toHostString(endpointMapping, key_source)
}

func applyOnUri(uri string, endpointMapping map[string]interface{}) string {
	parsedUrl := parseUri(uri)
	if sourceMatchesUri(parsedUrl, endpointMapping) {
		parsedUrl.Host = toHostString(endpointMapping, key_target)
	}
	return fmt.Sprintf("%v", parsedUrl)
}

func applyWriteReadUrlIfExists(credentials map[string]interface{}, endpointMapping map[string]interface{}, urlKey string) {
	if credentials[urlKey] != nil {
		credentials[urlKey] = applyOnReadWriteUrl(credentials[urlKey].(string), endpointMapping)
	}
}

func applyOnReadWriteUrl(url string, endpointMapping map[string]interface{}) string {
	sourcePort := fmt.Sprintf("%v", toStringMap(endpointMapping[key_source])[key_port])
	sourceHost := toStringMap(endpointMapping[key_source])[key_host].(string)
	sourceHostAndPort := toHostString(endpointMapping, key_source)
	targetString := toHostString(endpointMapping, key_target)

	newUrl := strings.Replace(url, sourceHostAndPort, targetString, 1)
	unchanged := (newUrl == url)
	if (sourcePort == default_postgres_port) && unchanged {
		newUrl = strings.Replace(url, sourceHost, targetString, 1)
	}
	return newUrl
}

func parseUri(uri string) *url.URL {
	url, _ := url.Parse(uri)
	return url
}

func sourceMatchesUri(url *url.URL, endpointMapping map[string]interface{}) bool {
	return url.Host == toHostString(endpointMapping, key_source)
}

func toHostString(endpointMapping map[string]interface{}, sourceOrTarget string) string {
	endpoint := toStringMap(endpointMapping[sourceOrTarget])
	return fmt.Sprintf("%v:%v", endpoint[key_host], endpoint[key_port])
}
