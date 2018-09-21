package model

import (
	"encoding/json"
	"net/url"
)

type Credentials struct {
	AdditionalProperties map[string]json.RawMessage
	Endpoints            []Endpoint
	Hostname             string
	Port                 int
	Uri                  string
	EndpointMappings     []EndpointMapping
}

type EndpointMapping struct {
	Source Endpoint
	Target Endpoint
}

func (credentials *Credentials) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &credentials.AdditionalProperties); err != nil {
		return err
	}
	err := removeProperty(credentials.AdditionalProperties, "end_points", &credentials.Endpoints)
	if err != nil {
		return err
	}
	err = removeProperty(credentials.AdditionalProperties, "hostname", &credentials.Hostname)
	if err != nil {
		return err
	}
	err = removeIntOrStringProperty(credentials.AdditionalProperties, "port", &credentials.Port)
	if err != nil {
		return err
	}
	err = removeProperty(credentials.AdditionalProperties, "uri", &credentials.Uri)
	if err != nil {
		return err
	}
	_, err = url.Parse(credentials.Uri)
	if err != nil {
		return err
	}
	err = removeProperty(credentials.AdditionalProperties, "endpoint_mappings", &credentials.EndpointMappings)
	if err != nil {
		return err
	}
	return nil
}

func (credentials Credentials) MarshalJSON() ([]byte, error) {
	properties := clone(credentials.AdditionalProperties)
	if len(credentials.Endpoints) != 0 {
		addProperty(properties, "end_points", credentials.Endpoints)
	}
	if len(credentials.Hostname) > 0 {
		addProperty(properties, "hostname", credentials.Hostname)
	}
	if len(credentials.Hostname) > 0 {
		addProperty(properties, "uri", credentials.Uri)
	}
	if credentials.Port != 0 {
		addProperty(properties, "port", credentials.Port)
	}
	if len(credentials.EndpointMappings) != 0 {
		addProperty(properties, "endpoint_mappings", credentials.EndpointMappings)
	}

	return json.Marshal(properties)
}
