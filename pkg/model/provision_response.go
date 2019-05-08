package model

//ProvisionResponse represents a provision request according to OSB-spec
type ProvisionResponse struct {
	AdditionalProperties additionalProperties
	NetworkProfiles []NetworkProfile
}



//UnmarshalJSON to ProvisionResponse
func (provisionResponse *ProvisionResponse) UnmarshalJSON(b []byte) error {
	return provisionResponse.AdditionalProperties.UnmarshalJSON(b, map[string]interface{}{"network_profiles": &provisionResponse.NetworkProfiles})
}

//MarshalJSON from ProvisionResponse
func (provisionResponse ProvisionResponse) MarshalJSON() ([]byte, error) {
	return provisionResponse.AdditionalProperties.MarshalJSON(map[string]interface{}{"network_profiles": &provisionResponse.NetworkProfiles})
}
