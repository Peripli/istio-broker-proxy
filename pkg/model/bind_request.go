package model

type BindRequest struct {
	AdditionalProperties AdditionalProperties
	NetworkData          NetworkDataRequest
}

type NetworkDataRequest struct {
	NetworkProfileID string      `json:"network_profile_id"`
	Data             DataRequest `json:"data"`
}

type DataRequest struct {
	ConsumerID string `json:"consumer_id"`
}

func (bindRequest *BindRequest) UnmarshalJSON(b []byte) error {
	return bindRequest.AdditionalProperties.UnmarshalJSON(b, map[string]interface{}{"network_data": &bindRequest.NetworkData})
}

func (bindRequest BindRequest) MarshalJSON() ([]byte, error) {
	return bindRequest.AdditionalProperties.MarshalJSON(map[string]interface{}{"network_data": &bindRequest.NetworkData})
}
