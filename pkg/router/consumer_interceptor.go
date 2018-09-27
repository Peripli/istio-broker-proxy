package router

import (
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/profiles"
)

type consumer_interceptor struct{}

func (c consumer_interceptor) preBind(request model.BindRequest) *model.BindRequest {
	request.NetworkData.Data.ConsumerId = ProxyConfiguration.ConsumerId
	request.NetworkData.NetworkProfileId = profiles.NetworkProfile
	return &request
}

func (c consumer_interceptor) postBind(request model.BindRequest, response model.BindResponse, bindId string) (*model.BindResponse, error) {
	return &response, nil
}
