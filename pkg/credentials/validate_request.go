package credentials

import (
	"encoding/json"
	"errors"
	"net/url"
)

func isValidUpdateRequestBody(request string) (bool, error) {
	var rawJson interface{}
	err := json.Unmarshal([]byte(request), &rawJson)
	if err != nil {
		err = errors.New("Error in unmarshalling: " + err.Error())
		return false, err
	}
	topLevelJson := rawJson.(map[string]interface{})
	_, err = hasField(topLevelJson, key_credentials)
	if err != nil {
		err = errors.New("Error in unmarshalling: " + err.Error())
		return false, err
	}
	_, err = hasField(topLevelJson, key_endpoint_mappings)
	if err != nil {
		err = errors.New("Error in unmarshalling: " + err.Error())
		return false, err
	}

	credentials := topLevelJson[key_credentials].(map[string]interface{})

	ok, err := isValidCredentials(credentials)
	if !ok {
		return false, err
	}

	endpointMappings := topLevelJson[key_endpoint_mappings]
	ok, err = isValidEndpointMappings(endpointMappings)
	if !ok {
		return false, err
	}

	for _, endpointMapping := range endpointMappings.([]interface{}) {
		ok = ok && shouldApply(toStringMap(toStringMap(endpointMapping)), credentials)
	}

	return ok, nil
}

func isValidEndpointMappings(endpointMappings interface{}) (bool, error) {
	switch endpointMappings.(type) {
	case []interface{}:
		if len(endpointMappings.([]interface{})) < 1 {
			err := errors.New("'endpoint_mappings' must contain at least one mapping")
			return false, err
		}
		for _, value := range endpointMappings.([]interface{}) {
			_, err := isValidEndpointMapping(value.(map[string]interface{}))

			if err != nil {
				return false, err
			}
		}
		return true, nil
	default:
		err := errors.New("'endpoint_mappings' must be an array")
		return false, err
	}
}

func isValidEndpointMapping(endpointMapping map[string]interface{}) (bool, error) {
	ok, err := hasField(endpointMapping, key_source)
	if !ok {
		return false, err
	}

	ok, err = hasField(endpointMapping, key_target)
	if !ok {
		return false, err
	}

	ok, err = isValidEndpoint(endpointMapping[key_source].(map[string]interface{}))
	if !ok {
		return false, err
	}

	ok, err = isValidEndpoint(endpointMapping[key_target].(map[string]interface{}))

	return ok, err
}

func isValidEndpoint(jsonMap map[string]interface{}) (bool, error) {
	ok, err := hasField(jsonMap, key_host)
	if !ok {
		return false, err
	}

	ok, err = hasField(jsonMap, key_port)
	return ok, err
}

func isValidCredentials(jsonMap map[string]interface{}) (bool, error) {
	ok, err := hasField(jsonMap, key_uri)
	if !ok {
		return false, err
	}

	ok, err = hasField(jsonMap, key_hostname)
	if !ok {
		return false, err
	}

	ok, err = hasField(jsonMap, key_port)
	if !ok {
		return false, err
	}

	ok, err = canParseUri(jsonMap[key_uri].(string))

	return ok, err
}

func canParseUri(uriValue string) (bool, error) {
	_, err := url.Parse(uriValue)
	return err == nil, err
}

func hasField(jsonMap map[string]interface{}, fieldName string) (bool, error) {
	if jsonMap[fieldName] == nil {
		err := errors.New("Invalid json, field not found: " + fieldName)
		return false, err
	}
	return true, nil
}

func shouldApply(endpoint map[string]interface{}, credentials map[string]interface{}) bool {
	url := parseUri(credentials[key_uri].(string))

	return sourceMatchesCredentials(credentials, endpoint) || sourceMatchesUri(url, endpoint)
}
