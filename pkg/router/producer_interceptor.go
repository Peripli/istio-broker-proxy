package router

import (
	"encoding/json"
	"github.com/Peripli/istio-broker-proxy/pkg/config"
	"github.com/Peripli/istio-broker-proxy/pkg/model"
	"github.com/Peripli/istio-broker-proxy/pkg/profiles"
	"log"
	"net/http"
)

//ProducerInterceptor contains config for the producer side
type ProducerInterceptor struct {
	LoadBalancerPort  int
	SystemDomain      string
	ProviderID        string
	IPAddress         string
	ServiceNamePrefix string
	PlanMetaData      string
	NetworkProfile    string
	ConfigStore       ConfigStore
}

//WriteIstioConfigFiles creates istio config for control plane route
func (c *ProducerInterceptor) WriteIstioConfigFiles(port int) error {
	return c.ConfigStore.CreateIstioConfig("istio-broker",
		config.CreateEntriesForExternalService("istio-broker", string(c.IPAddress), uint32(port), "istio-broker."+c.SystemDomain, "", 9000, c.ProviderID))
}

//PreBind see interface definition
func (c ProducerInterceptor) PreBind(request model.BindRequest) (*model.BindRequest, error) {
	// c.NetworkProfile (of provider) is already checked in main to be not empty
	consumerID := request.NetworkData.Data.ConsumerID
	if consumerID == "" {
		return nil, model.HTTPError{ErrorMsg: "InvalidConsumerID", Description: "no consumer ID included in bind request", StatusCode: http.StatusBadRequest}
	}
	return &request, nil
}

//PostBind see interface definition
func (c ProducerInterceptor) PostBind(request model.BindRequest, response model.BindResponse, bindingID string,
	adapt func(model.Credentials, []model.EndpointMapping) (*model.BindResponse, error)) (*model.BindResponse, error) {
	systemDomain := c.SystemDomain
	providerID := c.ProviderID
	if len(response.Endpoints) == 0 {
		response.Endpoints = response.Credentials.Endpoints
	}
	response.Credentials.Endpoints = nil
	profiles.AddIstioNetworkDataToResponse(providerID, bindingID, systemDomain, c.LoadBalancerPort, &response, c.NetworkProfile)

	err := c.writeIstioFilesForProvider(bindingID, &request, &response)
	if err != nil {
		c.PostDelete(bindingID)
		return nil, err
	}
	return &response, nil
}

//HasAdaptCredentials see interface definition
func (c ProducerInterceptor) HasAdaptCredentials() bool {
	return true
}

//PostDelete see interface definition
func (c ProducerInterceptor) PostDelete(bindID string) {
	err := c.ConfigStore.DeleteBinding(bindID)
	if err != nil {
		log.Printf("Ignoring error during removal of binding-id %s: %v\n", bindID, err)
	}
}

func (c ProducerInterceptor) writeIstioFilesForProvider(bindingID string, request *model.BindRequest, response *model.BindResponse) error {
	return c.ConfigStore.CreateIstioConfig(bindingID, config.CreateIstioConfigForProvider(request, response, bindingID, c.SystemDomain, c.ProviderID))
}

//PostCatalog see interface definition
func (c ProducerInterceptor) PostCatalog(catalog *model.Catalog) error {
	for i := range catalog.Services {
		catalog.Services[i].Name = c.ServiceNamePrefix + catalog.Services[i].Name
		if c.PlanMetaData != "" {
			for j := range catalog.Services[i].Plans {
				err := json.Unmarshal([]byte(c.PlanMetaData), &catalog.Services[i].Plans[j].MetaData)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
