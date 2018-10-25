package model

import (
	"fmt"
	"strings"
)

type RabbitMQCredentials struct {
	Credentials
	Hostname string
	Port     int
	Uri      string
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
	err := removeProperty(result.AdditionalProperties, uri_key, &result.Uri)
	if err != nil {
		return nil, err
	}
	if !(strings.HasPrefix(result.Uri, "amqp:")) {
		return nil, nil
	}
	err = removeProperties(result.AdditionalProperties, map[string]interface{}{
		hostname_key: &result.Hostname,
	})
	if err != nil {
		return nil, err
	}
	err = removeIntOrStringProperty(result.AdditionalProperties, port_key, &result.Port)
	if err != nil {
		return nil, err
	}
	if result.Uri == "" || result.Hostname == "" || result.Port == 0 {
		return nil, fmt.Errorf("Invalid rabbitmq credentials: %#v", result)
	}
	return &result, nil
}

func (credentials RabbitMQCredentials) ToCredentials() Credentials {
	result := Credentials{clone(credentials.AdditionalProperties), credentials.Endpoints}
	if len(credentials.Hostname) > 0 {
		addProperty(result.AdditionalProperties, hostname_key, credentials.Hostname)
	}
	if len(credentials.Uri) > 0 {
		addProperty(result.AdditionalProperties, uri_key, credentials.Uri)
	}

	if credentials.Port != 0 {
		addProperty(result.AdditionalProperties, port_key, credentials.Port)
	}
	return result
}

func (credentials *RabbitMQCredentials) Adapt(endpointMappings []EndpointMapping) {
	for _, endpointMapping := range endpointMappings {
		if credentials.Hostname == endpointMapping.Source.Host && credentials.Port == endpointMapping.Source.Port {
			credentials.Hostname = endpointMapping.Target.Host
			credentials.Port = endpointMapping.Target.Port
		}
		credentials.Uri = replaceInRabbitMqUrl(credentials.Uri, endpointMapping)

	}
}

func replaceInRabbitMqUrl(url string, endpointMapping EndpointMapping) string {
	return replaceInUrl(url, endpointMapping, 5672)
}
