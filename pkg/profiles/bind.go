package profiles

import (
	"encoding/json"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/endpoints"
)

type BindRequest struct {
	AdditionalProperties map[string]json.RawMessage
	NetworkData          NetworkDataRequest `json:"network_data"`
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
	NetworkData          NetworkDataResponse `json:"network_data"`
	Credentials          json.RawMessage     `json:"credentials"`
}

type NetworkDataResponse struct {
	NetworkProfileId string       `json:"network_profile_id"`
	Data             DataResponse `json:"data"`
}

type DataResponse struct {
	ProviderId string               `json:"provider_id"`
	Endpoints  []endpoints.Endpoint `json:"endpoints, omitempty"`
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
	err := addProperty(properties, "network_data", &bindRequest.NetworkData)
	if nil != err {
		return nil, err
	}
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
	return nil
}

func (bindResponse BindResponse) MarshalJSON() ([]byte, error) {
	properties := clone(bindResponse.AdditionalProperties)
	err := addProperty(properties, "network_data", &bindResponse.NetworkData)
	if nil != err {
		return nil, err
	}
	err = addProperty(properties, "credentials", &bindResponse.Credentials)
	if nil != err {
		return nil, err
	}
	return json.Marshal(properties)
}

func addProperty(additionalProperties map[string]json.RawMessage, key string, data interface{}) error {
	rawData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	additionalProperties[key] = json.RawMessage(rawData)

	return nil
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
