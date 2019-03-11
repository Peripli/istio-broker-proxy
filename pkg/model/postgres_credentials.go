package model

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	defaultPostgresPort = 5432
	hostnameKey         = "hostname"
	portKey             = "port"
	writeURLKey         = "write_url"
	readURLKey          = "read_url"
	uriKey              = "uri"
)

//PostgresCredentials contains credentials for a postgres
type PostgresCredentials struct {
	Credentials
	Hostname string
	Port     int
	URI      string
	WriteURL string
	ReadURL  string
}

//PostgresCredentialsConverter converts to postgresCredentials and adapts endpoints
func PostgresCredentialsConverter(credentials Credentials, endpointMappings []EndpointMapping) (*Credentials, error) {
	postgresCredentials, err := PostgresCredentialsFromCredentials(credentials)
	if err != nil {
		return nil, err
	}
	if postgresCredentials == nil {
		return nil, nil
	}
	postgresCredentials.adapt(endpointMappings)
	result := postgresCredentials.ToCredentials()
	return &result, nil
}

//PostgresCredentialsFromCredentials convert Credentials to PostgresCredentials
func PostgresCredentialsFromCredentials(credentials Credentials) (*PostgresCredentials, error) {
	result := PostgresCredentials{}
	result.Endpoints = credentials.Endpoints
	result.AdditionalProperties = clone(credentials.AdditionalProperties)
	err := removeProperty(result.AdditionalProperties, uriKey, &result.URI)
	if err != nil {
		return nil, err
	}
	if !(strings.HasPrefix(result.URI, "postgres:") || strings.HasPrefix(result.URI, "jdbc:postgres:")) {
		return nil, nil
	}
	err = removeProperties(result.AdditionalProperties, map[string]interface{}{
		hostnameKey: &result.Hostname,
		writeURLKey: &result.WriteURL,
		readURLKey:  &result.ReadURL,
	})
	if err != nil {
		return nil, err
	}
	err = removeIntOrStringProperty(result.AdditionalProperties, portKey, &result.Port)
	if err != nil {
		return nil, err
	}
	if result.URI == "" || result.Hostname == "" || result.Port == 0 {
		return nil, fmt.Errorf("Invalid postgres credentials: %#v", result)
	}
	return &result, nil
}

//ToCredentials converts to general Credentials
func (credentials PostgresCredentials) ToCredentials() Credentials {
	result := Credentials{clone(credentials.AdditionalProperties), credentials.Endpoints}
	if len(credentials.Hostname) > 0 {
		addProperty(result.AdditionalProperties, hostnameKey, credentials.Hostname)
	}
	if len(credentials.URI) > 0 {
		addProperty(result.AdditionalProperties, uriKey, credentials.URI)
	}
	if len(credentials.ReadURL) > 0 {
		addProperty(result.AdditionalProperties, readURLKey, credentials.ReadURL)
	}
	if len(credentials.WriteURL) > 0 {
		addProperty(result.AdditionalProperties, writeURLKey, credentials.WriteURL)
	}
	if credentials.Port != 0 {
		addProperty(result.AdditionalProperties, portKey, credentials.Port)
	}
	return result
}

func (credentials *PostgresCredentials) adapt(endpointMappings []EndpointMapping) {
	for _, endpointMapping := range endpointMappings {
		if credentials.Hostname == endpointMapping.Source.Host && credentials.Port == endpointMapping.Source.Port {
			credentials.Hostname = endpointMapping.Target.Host
			credentials.Port = endpointMapping.Target.Port
		}
		credentials.URI = replaceInPostgresURL(credentials.URI, endpointMapping)
		credentials.ReadURL = replaceInPostgresURL(credentials.ReadURL, endpointMapping)
		credentials.WriteURL = replaceInPostgresURL(credentials.WriteURL, endpointMapping)
	}
}

func replaceInPostgresURL(url string, endpointMapping EndpointMapping) string {
	return replaceInURL(url, endpointMapping, defaultPostgresPort)
}

func replaceInURL(url string, endpointMapping EndpointMapping, defaultPort int) string {
	pattern := toHostPortPattern(endpointMapping.Source, defaultPort)
	return pattern.ReplaceAllString(url, "${1}"+toHostString(endpointMapping.Target)+"${2}")
}

func toHostPortPattern(endpoint Endpoint, defaultPort int) *regexp.Regexp {
	// the groups are only there to capture the rest of the string around the endpoint in question
	if endpoint.Port == defaultPort {
		return regexp.MustCompile(fmt.Sprintf("(\\W)%s(?::%d|)(\\W)", regexp.QuoteMeta(endpoint.Host), endpoint.Port))
	}
	return regexp.MustCompile(fmt.Sprintf("(\\W)%s:%d(\\W|$)", regexp.QuoteMeta(endpoint.Host), endpoint.Port))
}

func toHostString(endpoint Endpoint) string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}
