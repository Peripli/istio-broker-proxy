package endpoints

import (
	"bytes"
	"encoding/json"
	"strings"
)

type responseData struct {
	Credentials map[string]interface{} `json:"credentials"`
	Endpoints   []map[string]string    `json:"endpoints"`
}

func isPostgres(data responseData) bool {
	if data.Credentials["uri"] == nil {
		return false
	}

	uri := data.Credentials["uri"].(string)
	isPostgres := strings.HasPrefix(uri, "postgres://")
	return isPostgres
}

func GenerateEndpoint(data []byte) ([]byte, error) {
	var input responseData
	err := json.NewDecoder(bytes.NewReader(data)).Decode(&input)
	if err != nil {
		return nil, err
	}

	if isPostgres(input) {
		jsonResult, err := generateEndpointForPostgres(input)
		return jsonResult, err
	}

	return data, nil
}

func generateEndpointForPostgres(input responseData) ([]byte, error) {
	endpoint := make(map[string]string)
	endpoint["port"] = input.Credentials["port"].(string)
	endpoint["hostname"] = input.Credentials["hostname"].(string)
	input.Endpoints = append(input.Endpoints, endpoint)
	jsonResult, err := json.Marshal(input)
	return jsonResult, err
}
