package model

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const (
	default_postgres_port = 5432
	hostname_key          = "hostname"
	port_key              = "port"
	write_url_key         = "write_url"
	read_url_key          = "read_url"
	uri_key               = "uri"
)

type PostgresCredentials struct {
	Credentials
	Hostname string
	Port     int
	Uri      string
	WriteUrl string
	ReadUrl  string
}

func PostgresCredentialsFromCredentials(credentials Credentials) (*PostgresCredentials, error) {
	result := PostgresCredentials{}
	result.Endpoints = credentials.Endpoints
	result.AdditionalProperties = clone(credentials.AdditionalProperties)
	err := removeProperty(result.AdditionalProperties, uri_key, &result.Uri)
	if err != nil {
		return nil, err
	}
	if !(strings.HasPrefix(result.Uri, "postgres:") || strings.HasPrefix(result.Uri, "jdbc:postgres:")) {
		return nil, nil
	}
	err = removeProperties(result.AdditionalProperties, map[string]interface{}{
		hostname_key:  &result.Hostname,
		write_url_key: &result.WriteUrl,
		read_url_key:  &result.ReadUrl,
	})
	if err != nil {
		return nil, err
	}
	err = removeIntOrStringProperty(result.AdditionalProperties, port_key, &result.Port)
	if err != nil {
		return nil, err
	}
	if result.Uri == "" || result.Hostname == "" || result.Port == 0 {
		return nil, fmt.Errorf("Invalid postgres credentials: %#v", result)
	}
	_, err = url.Parse(result.Uri)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (credentials PostgresCredentials) ToCredentials() Credentials {
	result := Credentials{clone(credentials.AdditionalProperties), credentials.Endpoints}
	if len(credentials.Hostname) > 0 {
		addProperty(result.AdditionalProperties, hostname_key, credentials.Hostname)
	}
	if len(credentials.Uri) > 0 {
		addProperty(result.AdditionalProperties, uri_key, credentials.Uri)
	}
	if len(credentials.ReadUrl) > 0 {
		addProperty(result.AdditionalProperties, read_url_key, credentials.ReadUrl)
	}
	if len(credentials.WriteUrl) > 0 {
		addProperty(result.AdditionalProperties, write_url_key, credentials.WriteUrl)
	}
	if credentials.Port != 0 {
		addProperty(result.AdditionalProperties, port_key, credentials.Port)
	}
	return result
}

func (credentials *PostgresCredentials) Adopt(endpointMappings []EndpointMapping) {
	for _, endpointMapping := range endpointMappings {
		if credentials.Hostname == endpointMapping.Source.Host && credentials.Port == endpointMapping.Source.Port {
			credentials.Hostname = endpointMapping.Target.Host
			credentials.Port = endpointMapping.Target.Port
		}
		credentials.Uri = replaceInUrl(credentials.Uri, endpointMapping)
		credentials.ReadUrl = replaceInUrl(credentials.ReadUrl, endpointMapping)
		credentials.WriteUrl = replaceInUrl(credentials.WriteUrl, endpointMapping)
	}
}

func replaceInUrl(url string, endpointMapping EndpointMapping) string {
	pattern := toHostPortPattern(endpointMapping.Source)
	return pattern.ReplaceAllString(url, "${1}"+toHostString(endpointMapping.Target)+"${2}")
}

func toHostPortPattern(endpoint Endpoint) *regexp.Regexp {
	// the groups are only there to capture the rest of the string around the endpoint in question
	if endpoint.Port == default_postgres_port {
		return regexp.MustCompile(fmt.Sprintf("(\\W)%s(?::%d|)(\\W)", regexp.QuoteMeta(endpoint.Host), endpoint.Port))
	}
	return regexp.MustCompile(fmt.Sprintf("(\\W)%s:%d(\\W)", regexp.QuoteMeta(endpoint.Host), endpoint.Port))
}

func toHostString(endpoint Endpoint) string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}
