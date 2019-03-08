package model

import (
	"fmt"
	"strings"
)

type RabbitMQCredentials struct {
	Credentials
	Hostname string
	Port     int
	URI      string
}

func RabbitMQCredentialsConverter(credentials Credentials, endpointMappings []EndpointMapping) (*Credentials, error) {
	rabbitMqCredentials, err := RabbitMQCredentialsFromCredentials(credentials)
	if err != nil {
		return nil, err
	}
	if rabbitMqCredentials == nil {
		return nil, nil
	}
	rabbitMqCredentials.Adapt(endpointMappings)
	result := rabbitMqCredentials.ToCredentials()
	return &result, nil
}

func RabbitMQCredentialsFromCredentials(credentials Credentials) (*RabbitMQCredentials, error) {
	result := RabbitMQCredentials{}
	result.Endpoints = credentials.Endpoints
	result.AdditionalProperties = clone(credentials.AdditionalProperties)
	err := removeProperty(result.AdditionalProperties, uriKey, &result.URI)
	if err != nil {
		return nil, err
	}
	if !(strings.HasPrefix(result.URI, "amqp:")) {
		return nil, nil
	}
	err = removeProperties(result.AdditionalProperties, map[string]interface{}{
		hostnameKey: &result.Hostname,
	})
	if err != nil {
		return nil, err
	}
	err = removeIntOrStringProperty(result.AdditionalProperties, portKey, &result.Port)
	if err != nil {
		return nil, err
	}
	if result.URI == "" || result.Hostname == "" || result.Port == 0 {
		return nil, fmt.Errorf("Invalid rabbitmq credentials: %#v", result)
	}
	return &result, nil
}

func (credentials RabbitMQCredentials) ToCredentials() Credentials {
	result := Credentials{clone(credentials.AdditionalProperties), credentials.Endpoints}
	if len(credentials.Hostname) > 0 {
		addProperty(result.AdditionalProperties, hostnameKey, credentials.Hostname)
	}
	if len(credentials.URI) > 0 {
		addProperty(result.AdditionalProperties, uriKey, credentials.URI)
	}

	if credentials.Port != 0 {
		addProperty(result.AdditionalProperties, portKey, credentials.Port)
	}
	return result
}

func (credentials *RabbitMQCredentials) Adapt(endpointMappings []EndpointMapping) {
	for _, endpointMapping := range endpointMappings {
		if credentials.Hostname == endpointMapping.Source.Host && credentials.Port == endpointMapping.Source.Port {
			credentials.Hostname = endpointMapping.Target.Host
			credentials.Port = endpointMapping.Target.Port
		}
		credentials.URI = replaceInRabbitMqURL(credentials.URI, endpointMapping)

	}
}

func replaceInRabbitMqURL(url string, endpointMapping EndpointMapping) string {
	return replaceInURL(url, endpointMapping, 5672)
}
