package model

import (
	"encoding/json"
	"fmt"
)

type Credentials struct {
	AdditionalProperties map[string]json.RawMessage
	Endpoints            []Endpoint
}

type BindRequest struct {
	AdditionalProperties map[string]json.RawMessage
	NetworkData          NetworkDataRequest
}

type NetworkDataRequest struct {
	NetworkProfileId string      `json:"network_profile_id"`
	Data             DataRequest `json:"data"`
}

type DataRequest struct {
	ConsumerId string `json:"consumer_id"`
}

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

func (bindRequest *BindRequest) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &bindRequest.AdditionalProperties); err != nil {
		return err
	}
	err := removeProperty(bindRequest.AdditionalProperties, "network_data", &bindRequest.NetworkData)
	if err != nil {
		return err
	}
	return nil
}

func (bindRequest BindRequest) MarshalJSON() ([]byte, error) {
	properties := clone(bindRequest.AdditionalProperties)
	addProperty(properties, "network_data", &bindRequest.NetworkData)
	return json.Marshal(properties)
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
	addProperty(properties, "network_data", &bindResponse.NetworkData)
	addProperty(properties, "credentials", &bindResponse.Credentials)
	if len(bindResponse.Endpoints) != 0 {

		addProperty(properties, "endpoints", bindResponse.Endpoints)
	}

	return json.Marshal(properties)
}

func (credentials *Credentials) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &credentials.AdditionalProperties); err != nil {
		return err
	}
	err := removeProperty(credentials.AdditionalProperties, "end_points", &credentials.Endpoints)
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

	return json.Marshal(properties)
}

func addProperty(additionalProperties map[string]json.RawMessage, key string, data interface{}) {
	rawData, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Sprintf("Error in marshal %v", err))
	}
	additionalProperties[key] = json.RawMessage(rawData)

}

func removeProperty(additionalProperties map[string]json.RawMessage, key string, data interface{}) error {
	rawData := additionalProperties[key]
	if rawData != nil {
		if err := json.Unmarshal(rawData, &data); err != nil {
			return err
		}
		delete(additionalProperties, key)
	}
	return nil
}

func clone(original map[string]json.RawMessage) map[string]json.RawMessage {
	copied := make(map[string]json.RawMessage)

	for key, value := range original {
		copied[key] = value
	}
	return copied
}
