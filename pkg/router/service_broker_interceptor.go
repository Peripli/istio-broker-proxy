package router

import (
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
)

type ServiceBrokerInterceptor interface {
	preBind(request model.BindRequest) *model.BindRequest
	postBind(request model.BindRequest, response model.BindResponse, bindId string) (*model.BindResponse, error)
	adaptCredentials(in []byte) ([]byte, error)
}

type NoOpInterceptor struct {
}

func (c NoOpInterceptor) preBind(request model.BindRequest) *model.BindRequest {
	return &request
}

func (c NoOpInterceptor) postBind(request model.BindRequest, response model.BindResponse, bindingId string) (*model.BindResponse, error) {
	return &response, nil
}

func (c NoOpInterceptor) adaptCredentials(in []byte) ([]byte, error) {
	return in, nil
}
