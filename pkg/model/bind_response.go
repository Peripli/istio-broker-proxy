package model

import (
	"encoding/json"
)

type BindResponse struct {
	AdditionalProperties map[string]json.RawMessage
	NetworkData          NetworkDataResponse
	Credentials          Credentials
	Endpoints            []Endpoint
}

type NetworkDataResponse struct {
	NetworkProfileId string       `json:"network_profile_id"`
	Data             DataResponse `json:"data"`
}

type DataResponse struct {
	ProviderId string     `json:"provider_id"`
	Endpoints  []Endpoint `json:"endpoints, omitempty"`
}

func (bindResponse *BindResponse) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &bindResponse.AdditionalProperties); err != nil {
		return err
	}
	err := removeProperty(bindResponse.AdditionalProperties, "network_data", &bindResponse.NetworkData)
	if err != nil {
		return err
	}
	err = removeProperty(bindResponse.AdditionalProperties, "credentials", &bindResponse.Credentials)
	if err != nil {
		return err
	}
	err = removeProperty(bindResponse.AdditionalProperties, "endpoints", &bindResponse.Endpoints)
	if err != nil {
		return err
	}
	return nil
}

func (bindResponse BindResponse) MarshalJSON() ([]byte, error) {
	properties := clone(bindResponse.AdditionalProperties)
	if len(bindResponse.NetworkData.NetworkProfileId) > 0 {
		addProperty(properties, "network_data", &bindResponse.NetworkData)
	}
	addProperty(properties, "credentials", &bindResponse.Credentials)
	if len(bindResponse.Endpoints) != 0 {
		addProperty(properties, "endpoints", bindResponse.Endpoints)
	}

	return json.Marshal(properties)
}
