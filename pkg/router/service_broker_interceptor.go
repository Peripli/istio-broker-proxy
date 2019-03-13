package router

import (
	"github.com/Peripli/istio-broker-proxy/pkg/model"
)

//ServiceBrokerInterceptor specifies methods to modify osb calls
type ServiceBrokerInterceptor interface {
	PreBind(request model.BindRequest) (*model.BindRequest, error)
	PostBind(request model.BindRequest, response model.BindResponse, bindID string,
		adapt func(model.Credentials, []model.EndpointMapping) (*model.BindResponse, error)) (*model.BindResponse, error)
	PostDelete(bindID string)
	PostCatalog(catalog *model.Catalog) error
	HasAdaptCredentials() bool
}

type noOpInterceptor struct {
}

func (c noOpInterceptor) PreBind(request model.BindRequest) (*model.BindRequest, error) {
	return &request, nil
}

func (c noOpInterceptor) PostBind(request model.BindRequest, response model.BindResponse, bindingID string,
	adapt func(model.Credentials, []model.EndpointMapping) (*model.BindResponse, error)) (*model.BindResponse, error) {
	return &response, nil
}

func (c noOpInterceptor) HasAdaptCredentials() bool {
	return false
}

func (c noOpInterceptor) PostDelete(bindID string) {
}

func (c noOpInterceptor) PostCatalog(catalog *model.Catalog) error {
	return nil
}

//NewNoOpInterceptor creates an empty ServiceBrokerInterceptor
func NewNoOpInterceptor() ServiceBrokerInterceptor {
	return &noOpInterceptor{}
}
