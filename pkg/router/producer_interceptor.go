package router

import (
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/profiles"
)

type producer_interceptor struct{}

func (c producer_interceptor) preBind(request model.BindRequest) *model.BindRequest {
	return &request
}

func (c producer_interceptor) postBind(request model.BindRequest, response model.BindResponse, bindingId string) (*model.BindResponse, error) {
	systemDomain := ProxyConfiguration.SystemDomain
	providerId := ProxyConfiguration.ProviderId
	if len(response.Endpoints) == 0 {
		response.Endpoints = response.Credentials.Endpoints
	}
	profiles.AddIstioNetworkDataToResponse(providerId, bindingId, systemDomain, ProxyConfiguration.LoadBalancerPort, &response)

	err := writeIstioFilesForProvider(ProxyConfiguration.IstioDirectory, bindingId, &request, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
