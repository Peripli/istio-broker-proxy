package model

//BindResponse represents a response to an OSB-bind
type BindResponse struct {
	AdditionalProperties additionalProperties
	NetworkData          NetworkDataResponse
	Credentials          Credentials
	Endpoints            []Endpoint
}

//NetworkDataResponse represents a osb-NetworkProfile field in a response
type NetworkDataResponse struct {
	NetworkProfileID string       `json:"network_profile_id"`
	Data             DataResponse `json:"data"`
}

//DataResponse represents the data section of an osb-NetworkProfile in a response
type DataResponse struct {
	ProviderID string     `json:"provider_id"`
	Endpoints  []Endpoint `json:"endpoints, omitempty"`
}

//UnmarshalJSON to a BindResponse
func (bindResponse *BindResponse) UnmarshalJSON(b []byte) error {
	return bindResponse.AdditionalProperties.UnmarshalJSON(b, map[string]interface{}{
		"network_data": &bindResponse.NetworkData,
		"credentials":  &bindResponse.Credentials,
		"endpoints":    &bindResponse.Endpoints,
	})
}

//MarshalJSON from BindResponse
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
