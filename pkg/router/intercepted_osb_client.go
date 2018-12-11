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

func (c *InterceptedOsbClient) Bind(instanceId string, bindingId string, bindRequest *model.BindRequest) (*model.BindResponse, error) {

	bindRequest = c.Interceptor.PreBind(*bindRequest)

	bindResponse, err := c.OsbClient.Bind(instanceId, bindingId, bindRequest)
	if err != nil {
		return nil, err
	}

	return c.Interceptor.PostBind(*bindRequest, *bindResponse, bindingId,
		func(credentials model.Credentials, mappings []model.EndpointMapping) (*model.BindResponse, error) {
			return c.OsbClient.AdaptCredentials(instanceId, bindingId, credentials, mappings)
		})

}

func (c *InterceptedOsbClient) Unbind(instanceId string, bindId string, rawQuery string) error {

	err := c.OsbClient.Unbind(instanceId, bindId, rawQuery)
	if err != nil {
		return err
	}
	return c.Interceptor.PostDelete(bindId)

}
