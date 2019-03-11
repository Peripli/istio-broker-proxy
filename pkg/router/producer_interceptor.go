package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Peripli/istio-broker-proxy/pkg/config"
	"github.com/Peripli/istio-broker-proxy/pkg/model"
	"github.com/Peripli/istio-broker-proxy/pkg/profiles"
	istioModel "istio.io/istio/pilot/pkg/model"
	"log"
	"os"
	"path"
)

//ProducerInterceptor contains config for the producer side
type ProducerInterceptor struct {
	LoadBalancerPort  int
	SystemDomain      string
	ProviderID        string
	IstioDirectory    string
	IPAddress         string
	ServiceNamePrefix string
	PlanMetaData      string
	NetworkProfile    string
}

func (c *ProducerInterceptor) WriteIstioConfigFiles(port int) error {
	return c.writeIstioConfigFiles("istio-broker",
		config.CreateEntriesForExternalService("istio-broker", string(c.IPAddress), uint32(port), "istio-broker."+c.SystemDomain, "", 9000, c.ProviderID))
}

func (c ProducerInterceptor) PreBind(request model.BindRequest) (*model.BindRequest, error) {
	return &request, nil
}

func (c ProducerInterceptor) PostBind(request model.BindRequest, response model.BindResponse, bindingID string,
	adapt func(model.Credentials, []model.EndpointMapping) (*model.BindResponse, error)) (*model.BindResponse, error) {
	if c.NetworkProfile == "" {
		return nil, errors.New("network profile not configured")
	}
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

func (c ProducerInterceptor) HasAdaptCredentials() bool {
	return true
}

func (c ProducerInterceptor) PostDelete(bindID string) error {
	fileName := path.Join(c.IstioDirectory, bindID) + ".yml"
	err := os.Remove(fileName)
	if err != nil {
		log.Printf("Ignoring error during removal of file %s: %v\n", fileName, err)
	}
	return nil
}

func (c ProducerInterceptor) writeIstioFilesForProvider(bindingID string, request *model.BindRequest, response *model.BindResponse) error {
	return c.writeIstioConfigFiles(bindingID, config.CreateIstioConfigForProvider(request, response, bindingID, c.SystemDomain, c.ProviderID))
}

func (c ProducerInterceptor) writeIstioConfigFiles(fileName string, configuration []istioModel.Config) error {
	ymlPath := path.Join(c.IstioDirectory, fileName) + ".yml"
	log.Printf("PATH to istio config: %v\n", ymlPath)
	file, err := os.Create(ymlPath)
	if nil != err {
		return fmt.Errorf("Unable to write istio configuration to file '%s': %s", fileName, err.Error())
	}
	defer file.Close()

	fileContent, err := config.ToYamlDocuments(configuration)
	if nil != err {
		return err
	}
	_, err = file.Write([]byte(fileContent))
	if nil != err {
		return err
	}
	return nil
}

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
