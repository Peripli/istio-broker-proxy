package model

//Credentials to be used in OSB-calls (credentials are a free-form hash)
type Credentials struct {
	AdditionalProperties additionalProperties
	Endpoints            []Endpoint
}

//EndpointMapping represents an endpoint mapping used for an adapt_credentials call according to osb-spec
type EndpointMapping struct {
	Source Endpoint
	Target Endpoint
}

//UnmarshalJSON to Credentials
func (credentials *Credentials) UnmarshalJSON(b []byte) error {
	return credentials.AdditionalProperties.UnmarshalJSON(b, map[string]interface{}{"end_points": &credentials.Endpoints})
}

//MarshalJSON from Credentials to bytes
func (credentials Credentials) MarshalJSON() ([]byte, error) {
	return credentials.AdditionalProperties.MarshalJSON(map[string]interface{}{"end_points": credentials.Endpoints})
}
