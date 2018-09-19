package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const key_uri string = "uri"

type Endpoint struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func (ep *Endpoint) UnmarshalJSON(b []byte) error {
	var untyped struct {
		Host string      `json:"host"`
		Port interface{} `json:"port"`
	}
	err := json.Unmarshal(b, &untyped)
	if err != nil {
		return err
	}

	ep.Host = untyped.Host
	if untyped.Port != nil {
		portAsString := fmt.Sprintf("%v", untyped.Port)
		ep.Port, err = strconv.Atoi(portAsString)
	}
	return err
}

type responseData struct {
	Credentials map[string]interface{} `json:"credentials"`
	Endpoints   []Endpoint             `json:"endpoints"`
}

func isPostgres(data responseData) bool {
	if data.Credentials[key_uri] == nil {
		return false
	}

	uri := data.Credentials[key_uri].(string)
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
	data, err := json.Marshal(input.Credentials["end_points"])
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &input.Endpoints)
	if err != nil {
		return nil, err
	}

	jsonResult, err := json.Marshal(input)
	return jsonResult, err
}
