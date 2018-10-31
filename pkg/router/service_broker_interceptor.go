package router

import (
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
)

type ServiceBrokerInterceptor interface {
	preBind(request model.BindRequest) *model.BindRequest
	postBind(request model.BindRequest, response model.BindResponse, bindId string,
		adapt func(model.Credentials, []model.EndpointMapping) (*model.BindResponse, error)) (*model.BindResponse, error)
	postDelete(bindId string) error
	postCatalog(catalog *model.Catalog)
	hasAdaptCredentials() bool
}

type NoOpInterceptor struct {
}

func (c NoOpInterceptor) preBind(request model.BindRequest) *model.BindRequest {
	return &request
}

func (c NoOpInterceptor) postBind(request model.BindRequest, response model.BindResponse, bindingId string,
	adapt func(model.Credentials, []model.EndpointMapping) (*model.BindResponse, error)) (*model.BindResponse, error) {
	return &response, nil
}

func (c NoOpInterceptor) hasAdaptCredentials() bool {
	return false
}

func (c NoOpInterceptor) postDelete(bindId string) error {
	return nil
}

func (c NoOpInterceptor) postCatalog(catalog *model.Catalog) {
}
