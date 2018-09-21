package credentials

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

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
		applyWriteReadUrlIfExists(credentials, endpointMapping)
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

func sourceMatchesReadWriteUrl(credentials map[string]interface{}, endpointMapping map[string]interface{}) bool {
	for _, key := range []string{key_read_url, key_write_url} {
		if credentials[key] != nil {
			if credentials[key] != replaceInUrl(credentials[key].(string), endpointMapping) {
				return true
			}
		}
	}
	return false
}

func applyWriteReadUrlIfExists(credentials map[string]interface{}, endpointMapping map[string]interface{}) {
	for _, key := range []string{key_read_url, key_write_url} {
		if credentials[key] != nil {
			credentials[key] = replaceInUrl(credentials[key].(string), endpointMapping)
		}
	}
}

func replaceInUrl(url string, endpointMapping map[string]interface{}) string {
	sourceHost := toStringMap(endpointMapping[key_source])[key_host].(string)
	sourcePort := fmt.Sprintf("%v", toStringMap(endpointMapping[key_source])[key_port])
	pattern := toHostPortPattern(sourceHost, sourcePort)
	newUrl := pattern.ReplaceAllString(url, "${1}"+toHostString(endpointMapping, key_target)+"${2}")
	if newUrl == url && sourcePort == default_postgres_port {
		pattern = toHostPattern(sourceHost)
		newUrl = pattern.ReplaceAllString(url, "${1}"+toHostString(endpointMapping, key_target)+"${2}")
	}
	return newUrl
}

func toHostPortPattern(host string, port string) *regexp.Regexp {
	// the groups are only there to capture the rest of the string around the endpoint in question
	pattern := fmt.Sprintf("^(.*\\W)%s:%s(\\W.*)$", strings.Replace(host, ".", "\\.", -1), port)

	return regexp.MustCompile(pattern)
}

func toHostPattern(host string) *regexp.Regexp {
	// the groups are only there to capture the rest of the string around the endpoint in question
	pattern := fmt.Sprintf("^(.*\\W)%s(\\W.*)$", strings.Replace(host, ".", "\\.", -1))

	return regexp.MustCompile(pattern)
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
