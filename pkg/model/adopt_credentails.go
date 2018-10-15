package model

type AdaptCredentialsRequest struct {
	Credentials      Credentials       `json:"credentials"`
	EndpointMappings []EndpointMapping `json:"endpoint_mappings"`
}

func Adapt(credentials Credentials, endpointMappings []EndpointMapping) (*BindResponse, error) {

	postgresCredentials, err := PostgresCredentialsFromCredentials(credentials)
	if err != nil {
		return nil, err
	}
	result := BindResponse{}
	if postgresCredentials == nil {
		result.Credentials = credentials
	} else {
		postgresCredentials.Adapt(endpointMappings)
		result.Credentials = postgresCredentials.ToCredentials()
	}

	for _, endpointMapping := range endpointMappings {
		result.Endpoints = append(result.Endpoints, endpointMapping.Target)
	}
	return &result, nil

}
