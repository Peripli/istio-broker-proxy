package router

import (
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
)

type ServiceBrokerInterceptor interface {
	preBind(request model.BindRequest) *model.BindRequest
	postBind(request model.BindRequest, response model.BindResponse, bindId string) (*model.BindResponse, error)
}

type noOpInterceptor struct {
}

func (c noOpInterceptor) preBind(request model.BindRequest) *model.BindRequest {
	return &request
}

func (c noOpInterceptor) postBind(request model.BindRequest, response model.BindResponse, bindingId string) (*model.BindResponse, error) {
	return &response, nil
}
