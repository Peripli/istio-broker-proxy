package router

import (
	"github.com/Peripli/istio-broker-proxy/pkg/model"
)

type InterceptedOsbClient struct {
	OsbClient   *OsbClient
	Interceptor ServiceBrokerInterceptor
}

func (c *InterceptedOsbClient) GetCatalog() (*model.Catalog, error) {
	catalog, err := c.OsbClient.GetCatalog()
	if err != nil {
		return nil, err
	}
	err = c.Interceptor.PostCatalog(catalog)
	return catalog, err
}

func (c *InterceptedOsbClient) Bind(bindingID string, bindRequest *model.BindRequest) (*model.BindResponse, error) {
	bindRequest, err := c.Interceptor.PreBind(*bindRequest)
	if err != nil {
		return nil, err
	}

	bindResponse, err := c.OsbClient.Bind(bindRequest)
	if err != nil {
		return nil, err
	}

	return c.Interceptor.PostBind(*bindRequest, *bindResponse, bindingID,
		func(credentials model.Credentials, mappings []model.EndpointMapping) (*model.BindResponse, error) {
			return c.OsbClient.AdaptCredentials(credentials, mappings)
		})
}

func (c *InterceptedOsbClient) Unbind(bindID string) error {
	err := c.OsbClient.Unbind()
	if err != nil {
		return err
	}
	return c.Interceptor.PostDelete(bindID)
}
