package model

import "encoding/json"

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
	if len(bindRequest.NetworkData.NetworkProfileId) > 0 {
		addProperty(properties, "network_data", &bindRequest.NetworkData)
	}
	return json.Marshal(properties)
}
