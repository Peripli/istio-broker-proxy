package profiles

import (
	"fmt"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
)

const network_profile = "urn:com.sap.istio:public"

func AddIstioNetworkDataToResponse(providerId string, serviceId string, systemDomain string, portNumber int, body *model.BindResponse) {

	endpointCount := len(body.Endpoints)

	endpointHosts := createEndpointHostsBasedOnSystemDomainServiceId(serviceId, systemDomain, endpointCount)

	newEndpoints := make([]model.Endpoint, 0)
	for _, endpointHost := range endpointHosts {

		newEndpoints = append(newEndpoints, model.Endpoint{
			endpointHost,
			portNumber,
		},
		)
	}

	body.NetworkData.NetworkProfileId = network_profile
	body.NetworkData.Data.ProviderId = providerId
	body.NetworkData.Data.Endpoints = newEndpoints

}

func createEndpointHostsBasedOnSystemDomainServiceId(serviceId string, systemDomain string, count int) []string {
	var endpointsHosts []string

	for i := 0; i < count; i++ {
		newHost := fmt.Sprintf("%d.%s.%s", i+1, serviceId, systemDomain)
		endpointsHosts = append(endpointsHosts, newHost)
	}
	return endpointsHosts
}
