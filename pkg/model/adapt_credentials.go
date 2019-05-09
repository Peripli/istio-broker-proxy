package model

import (
	"errors"
)

//AdaptCredentialsRequest in accordance with OSB-spec
type AdaptCredentialsRequest struct {
	Credentials      Credentials       `json:"credentials"`
	EndpointMappings []EndpointMapping `json:"endpoint_mappings"`
}

type credentialConverter func(credentials Credentials, endpointMappings []EndpointMapping) (*Credentials, error)

var converters = []credentialConverter{
	PostgresCredentialsConverter,
	RabbitMQCredentialsConverter,
	func(credentials Credentials, endpointMappings []EndpointMapping) (*Credentials, error) {
		return &credentials, nil
	},
}

//Adapt credentials according to the specified EndpointMapping
func Adapt(credentials Credentials, endpointMappings []EndpointMapping) (*BindResponse, error) {

	if len(endpointMappings) == 0 {
		return nil, errors.New("No endpoint mappings available")
	}
	result := BindResponse{}
	for _, converter := range converters {
		c, err := converter(credentials, endpointMappings)
		if err != nil {
			return nil, err
		}
		if c != nil {
			result.Credentials = *c
			break
		}
	}

	for _, endpointMapping := range endpointMappings {
		result.Endpoints = append(result.Endpoints, endpointMapping.Target)
	}
	return &result, nil

}
