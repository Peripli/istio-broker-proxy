package model

type AdpotCredentialsRequest struct {
	Credentials      Credentials       `json:"credentials"`
	EndpointMappings []EndpointMapping `json:"endpoint_mappings"`
}

func Adopt(request AdpotCredentialsRequest) (*BindResponse, error) {

	postgresCredentials, err := PostgresCredentialsFromCredentials(request.Credentials)
	if err != nil {
		return nil, err
	}
	result := BindResponse{}
	if postgresCredentials == nil {
		result.Credentials = request.Credentials
	} else {
		postgresCredentials.Adopt(request.EndpointMappings)
		result.Credentials = postgresCredentials.ToCredentials()
	}

	for _, endpointMapping := range request.EndpointMappings {
		result.Endpoints = append(result.Endpoints, endpointMapping.Target)
	}
	return &result, nil

}
