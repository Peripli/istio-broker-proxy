package model

type BindRequest struct {
	AdditionalProperties additionalProperties
	NetworkData          NetworkDataRequest
}

//NetworkDataRequest represents a osb-NetworkProfile field in a request
type NetworkDataRequest struct {
	NetworkProfileID string      `json:"network_profile_id"`
	Data             DataRequest `json:"data"`
}

//DataRequest represents the data section of an osb-NetworkProfile
type DataRequest struct {
	ConsumerID string `json:"consumer_id"`
}

func (bindRequest *BindRequest) UnmarshalJSON(b []byte) error {
	return bindRequest.AdditionalProperties.UnmarshalJSON(b, map[string]interface{}{"network_data": &bindRequest.NetworkData})
}

func (bindRequest BindRequest) MarshalJSON() ([]byte, error) {
	return bindRequest.AdditionalProperties.MarshalJSON(map[string]interface{}{"network_data": &bindRequest.NetworkData})
}
