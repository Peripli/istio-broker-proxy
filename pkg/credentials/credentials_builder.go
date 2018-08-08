package credentials

import "encoding/json"

type credentialBuilder struct {
	credentials     map[string]interface{}
	endpointMapping []map[string]map[string]interface{}
}

func newCredentialBuilder() *credentialBuilder {
	var b credentialBuilder
	b.credentials = make(map[string]interface{})

	b.setHostPort("unused", "9999")
	b.setUri("unused")
	return &b
}

func (b credentialBuilder) build() string {
	object := make(map[string]interface{})
	object["credentials"] = b.credentials
	object["endpoint_mappings"] = b.endpointMapping

	asString, _ := json.Marshal(object)

	return string(asString)
}

func (b *credentialBuilder) setHostPort(hostname string, port interface{}) *credentialBuilder {
	b.credentials["hostname"] = hostname
	b.credentials["port"] = port
	return b
}

func (b *credentialBuilder) setUri(uri string) *credentialBuilder {
	b.credentials["uri"] = uri
	return b
}

func (b *credentialBuilder) setEndpointMapping(sourceHost string, sourcePort interface{}, targetHost string, targetPort interface{}) *credentialBuilder {
	b.endpointMapping = nil
	return b.addEndpointMapping(sourceHost, sourcePort, targetHost, targetPort)
}

func (b *credentialBuilder) addEndpointMapping(sourceHost string, sourcePort interface{}, targetHost string, targetPort interface{}) *credentialBuilder {
	entry := make(map[string]map[string]interface{})

	entry["source"] = make(map[string]interface{})
	entry["source"]["host"] = sourceHost
	entry["source"]["port"] = sourcePort

	entry["target"] = make(map[string]interface{})
	entry["target"]["host"] = targetHost
	entry["target"]["port"] = targetPort

	b.endpointMapping = append(b.endpointMapping, entry)
	return b
}
