package model

type BindResponse struct {
	AdditionalProperties additionalProperties
	NetworkData          NetworkDataResponse
	Credentials          Credentials
	Endpoints            []Endpoint
}

type NetworkDataResponse struct {
	NetworkProfileID string       `json:"network_profile_id"`
	Data             DataResponse `json:"data"`
}

type DataResponse struct {
	ProviderID string     `json:"provider_id"`
	Endpoints  []Endpoint `json:"endpoints, omitempty"`
}

func (bindResponse *BindResponse) UnmarshalJSON(b []byte) error {
	return bindResponse.AdditionalProperties.UnmarshalJSON(b, map[string]interface{}{
		"network_data": &bindResponse.NetworkData,
		"credentials":  &bindResponse.Credentials,
		"endpoints":    &bindResponse.Endpoints,
	})
}

func (bindResponse BindResponse) MarshalJSON() ([]byte, error) {
	mapping := map[string]interface{}{
		"credentials": &bindResponse.Credentials,
		"endpoints":   bindResponse.Endpoints,
	}
	if len(bindResponse.NetworkData.NetworkProfileID) > 0 || len(bindResponse.NetworkData.Data.Endpoints) > 0 {
		mapping["network_data"] = &bindResponse.NetworkData
	}
	return bindResponse.AdditionalProperties.MarshalJSON(mapping)
}
