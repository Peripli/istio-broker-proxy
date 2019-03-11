package router

import (
	"github.com/Peripli/istio-broker-proxy/pkg/model"
)

type interceptedOsbClient struct {
	OsbClient   *osbClient
	Interceptor ServiceBrokerInterceptor
}

func (c *interceptedOsbClient) GetCatalog() (*model.Catalog, error) {
	catalog, err := c.OsbClient.getCatalog()
	if err != nil {
		return nil, err
	}
	err = c.Interceptor.PostCatalog(catalog)
	return catalog, err
}

func (c *interceptedOsbClient) Bind(bindingID string, bindRequest *model.BindRequest) (*model.BindResponse, error) {
	bindRequest, err := c.Interceptor.PreBind(*bindRequest)
	if err != nil {
		return nil, err
	}

	bindResponse, err := c.OsbClient.bind(bindRequest)
	if err != nil {
		return nil, err
	}

	return c.Interceptor.PostBind(*bindRequest, *bindResponse, bindingID,
		func(credentials model.Credentials, mappings []model.EndpointMapping) (*model.BindResponse, error) {
			return c.OsbClient.adaptCredentials(credentials, mappings)
		})
}

func (c *interceptedOsbClient) Unbind(bindID string) error {
	err := c.OsbClient.unbind()
	if err != nil {
		return err
	}
	return c.Interceptor.PostDelete(bindID)
}
