package router

import (
	"github.com/Peripli/istio-broker-proxy/pkg/model"
)

type ServiceBrokerInterceptor interface {
	PreBind(request model.BindRequest) (*model.BindRequest, error)
	PostBind(request model.BindRequest, response model.BindResponse, bindID string,
		adapt func(model.Credentials, []model.EndpointMapping) (*model.BindResponse, error)) (*model.BindResponse, error)
	PostDelete(bindID string) error
	PostCatalog(catalog *model.Catalog) error
	HasAdaptCredentials() bool
}

type NoOpInterceptor struct {
}

func (c NoOpInterceptor) PreBind(request model.BindRequest) (*model.BindRequest, error) {
	return &request, nil
}

func (c NoOpInterceptor) PostBind(request model.BindRequest, response model.BindResponse, bindingID string,
	adapt func(model.Credentials, []model.EndpointMapping) (*model.BindResponse, error)) (*model.BindResponse, error) {
	return &response, nil
}

func (c NoOpInterceptor) HasAdaptCredentials() bool {
	return false
}

func (c NoOpInterceptor) PostDelete(bindID string) error {
	return nil
}

func (c NoOpInterceptor) PostCatalog(catalog *model.Catalog) error {
	return nil
}
