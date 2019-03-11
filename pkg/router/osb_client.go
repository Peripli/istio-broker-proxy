package router

import (
	"github.com/Peripli/istio-broker-proxy/pkg/api"
	"github.com/Peripli/istio-broker-proxy/pkg/model"
)

type OsbClient struct {
	api.RestClient
}

func (client *OsbClient) AdaptCredentials(credentials model.Credentials, mapping []model.EndpointMapping) (*model.BindResponse, error) {
	var bindResponse model.BindResponse
	err := client.Post(&model.AdaptCredentialsRequest{Credentials: credentials, EndpointMappings: mapping}).
		AppendPath("/adapt_credentials").
		Do().
		Into(&bindResponse)
	return &bindResponse, err
}

func (client *OsbClient) GetCatalog() (*model.Catalog, error) {
	var catalog model.Catalog
	err := client.Get().
		Do().
		Into(&catalog)
	return &catalog, err
}

func (client *OsbClient) Bind(request *model.BindRequest) (*model.BindResponse, error) {
	var bindResponse model.BindResponse

	err := client.Put(request).
		Do().
		Into(&bindResponse)

	return &bindResponse, err
}

func (client *OsbClient) Unbind() error {
	return client.Delete().
		Do().
		Error()
}
