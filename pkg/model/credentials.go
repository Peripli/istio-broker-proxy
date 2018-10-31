package model

type Credentials struct {
	AdditionalProperties AdditionalProperties
	Endpoints            []Endpoint
}

type EndpointMapping struct {
	Source Endpoint
	Target Endpoint
}

func (credentials *Credentials) UnmarshalJSON(b []byte) error {
	return credentials.AdditionalProperties.UnmarshalJSON(b, map[string]interface{}{"end_points": &credentials.Endpoints})
}

func (credentials Credentials) MarshalJSON() ([]byte, error) {
	return credentials.AdditionalProperties.MarshalJSON(map[string]interface{}{"end_points": credentials.Endpoints})
}
