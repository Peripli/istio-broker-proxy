package profiles

import (
	"fmt"
	"github.com/Peripli/istio-broker-proxy/pkg/model"
)

func AddIstioNetworkDataToResponse(providerID string, bindingID string, systemDomain string, portNumber int, body *model.BindResponse, networkProfile string) {

	endpointCount := len(body.Endpoints)

	endpointHosts := createEndpointHostsBasedOnSystemDomainServiceID(bindingID, systemDomain, endpointCount)

	newEndpoints := make([]model.Endpoint, 0)
	for _, endpointHost := range endpointHosts {

		newEndpoints = append(newEndpoints, model.Endpoint{
			endpointHost,
			portNumber,
		},
		)
	}

	body.NetworkData.NetworkProfileID = networkProfile
	body.NetworkData.Data.ProviderID = providerID
	body.NetworkData.Data.Endpoints = newEndpoints

}

func createEndpointHostsBasedOnSystemDomainServiceID(bindingID string, systemDomain string, count int) []string {
	var endpointsHosts []string

	for i := 0; i < count; i++ {
		newHost := CreateEndpointHosts(bindingID, systemDomain, i)
		endpointsHosts = append(endpointsHosts, newHost)
	}
	return endpointsHosts
}

func CreateEndpointHosts(bindingID string, systemDomain string, index int) string {
	return fmt.Sprintf("%d.%s.%s", index, bindingID, systemDomain)
}
