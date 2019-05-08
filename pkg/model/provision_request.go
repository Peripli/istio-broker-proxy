package model

import "encoding/json"

//ProvisionRequest represents a provision request according to OSB-spec
type ProvisionRequest struct {
	AdditionalProperties additionalProperties
	NetworkProfiles []NetworkProfile
}

//NetworkProfile represents the network profile in provision or bind calls according to OSB-spec
type NetworkProfile struct {
	ID string      `json:"id"`
	Data           json.RawMessage `json:"data"`
}
//UnmarshalJSON to ProvisionRequest
func (provisionRequest *ProvisionRequest) UnmarshalJSON(b []byte) error {
	return provisionRequest.AdditionalProperties.UnmarshalJSON(b, map[string]interface{}{"network_profiles": &provisionRequest.NetworkProfiles})
}

//MarshalJSON from ProvisionRequest
func (provisionRequest ProvisionRequest) MarshalJSON() ([]byte, error) {
	return provisionRequest.AdditionalProperties.MarshalJSON(map[string]interface{}{"network_profiles": &provisionRequest.NetworkProfiles})
}
