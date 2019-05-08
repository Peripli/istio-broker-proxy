package router

import (
	"github.com/Peripli/istio-broker-proxy/pkg/model"
)

//ServiceBrokerInterceptor specifies methods to modify osb calls
type ServiceBrokerInterceptor interface {
	PreProvision(request model.ProvisionRequest) (*model.ProvisionRequest, error)
	PostProvision(request model.ProvisionRequest, response model.ProvisionResponse) (*model.ProvisionResponse, error)
	PreBind(request model.BindRequest) (*model.BindRequest, error)
	PostBind(request model.BindRequest, response model.BindResponse, bindID string,
		adapt func(model.Credentials, []model.EndpointMapping) (*model.BindResponse, error)) (*model.BindResponse, error)
	PostUnbind(bindID string)
	PostCatalog(catalog *model.Catalog) error
	HasAdaptCredentials() bool
}

type noOpInterceptor struct {
}

func (c noOpInterceptor) PreProvision(request model.ProvisionRequest) (*model.ProvisionRequest, error) {
	return &request,nil
}

func (c noOpInterceptor) PostProvision(request model.ProvisionRequest, response model.ProvisionResponse) (*model.ProvisionResponse, error) {
	return &response, nil
}

var _ ServiceBrokerInterceptor = &noOpInterceptor{}


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

func (c noOpInterceptor) PostUnbind(bindID string) {
}

func (c noOpInterceptor) PostCatalog(catalog *model.Catalog) error {
	return nil
}

//NewNoOpInterceptor creates an empty ServiceBrokerInterceptor
func NewNoOpInterceptor() ServiceBrokerInterceptor {
	return &noOpInterceptor{}
}
