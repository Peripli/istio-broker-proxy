package router

import (
	"fmt"
	"github.com/Peripli/istio-broker-proxy/pkg/model"
)

type OsbClient struct {
	RestClient
}

func (client *OsbClient) AdaptCredentials(instanceId string, bindId string, credentials model.Credentials, mapping []model.EndpointMapping) (*model.BindResponse, error) {

	var bindResponse model.BindResponse
	err := client.Post(&model.AdaptCredentialsRequest{Credentials: credentials, EndpointMappings: mapping}).
		Path(fmt.Sprintf("v2/service_instances/%s/service_bindings/%s/adapt_credentials", instanceId, bindId)).
		Do().
		Into(&bindResponse)
	return &bindResponse, err

}

func (client *OsbClient) GetCatalog() (*model.Catalog, error) {
	var catalog model.Catalog
	err := client.Get().
		Path("v2/catalog").
		Do().
		Into(&catalog)
	return &catalog, err

}

func (client *OsbClient) Bind(instanceId string, bindId string, request *model.BindRequest) (*model.BindResponse, error) {
	var bindResponse model.BindResponse

	err := client.Put(request).
		Path(fmt.Sprintf("v2/service_instances/%s/service_bindings/%s", instanceId, bindId)).
		Do().
		Into(&bindResponse)

	return &bindResponse, err

}

func (client *OsbClient) Unbind(instanceId string, bindId string, rawQuery string) error {

	return client.Delete().
		Path(fmt.Sprintf("v2/service_instances/%s/service_bindings/%s?%s", instanceId, bindId, rawQuery)).
		Do().
		Error()

}
