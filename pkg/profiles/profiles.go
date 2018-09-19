package profiles

import (
	"fmt"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
)

const NetworkProfile = "urn:com.sap.istio:public"

func AddIstioNetworkDataToResponse(providerId string, bindingId string, systemDomain string, portNumber int, body *model.BindResponse) {

	endpointCount := len(body.Endpoints)

	endpointHosts := createEndpointHostsBasedOnSystemDomainServiceId(bindingId, systemDomain, endpointCount)

	newEndpoints := make([]model.Endpoint, 0)
	for _, endpointHost := range endpointHosts {

		newEndpoints = append(newEndpoints, model.Endpoint{
			endpointHost,
			portNumber,
		},
		)
	}

	body.NetworkData.NetworkProfileId = NetworkProfile
	body.NetworkData.Data.ProviderId = providerId
	body.NetworkData.Data.Endpoints = newEndpoints

}

func createEndpointHostsBasedOnSystemDomainServiceId(bindingId string, systemDomain string, count int) []string {
	var endpointsHosts []string

	for i := 0; i < count; i++ {
		newHost := CreateEndpointHosts(bindingId, systemDomain, i)
		endpointsHosts = append(endpointsHosts, newHost)
	}
	return endpointsHosts
}

func CreateEndpointHosts(bindingId string, systemDomain string, index int) string {
	return fmt.Sprintf("%d.%s.%s", index, bindingId, systemDomain)
}
