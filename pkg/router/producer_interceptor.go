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

type ProducerInterceptor struct {
	LoadBalancerPort  int
	SystemDomain      string
	ProviderId        string
	IstioDirectory    string
	IpAddress         string
	ServiceNamePrefix string
	PlanMetaData      string
	NetworkProfile    string
}

func (c *ProducerInterceptor) WriteIstioConfigFiles(port int) error {
	return c.writeIstioConfigFiles("istio-broker",
		config.CreateEntriesForExternalService("istio-broker", string(c.IpAddress), uint32(port), "istio-broker."+c.SystemDomain, "", 9000, c.ProviderId))
}

func (c ProducerInterceptor) PreBind(request model.BindRequest) (*model.BindRequest, error) {
	return &request, nil
}

func (c ProducerInterceptor) PostBind(request model.BindRequest, response model.BindResponse, bindingId string,
	adapt func(model.Credentials, []model.EndpointMapping) (*model.BindResponse, error)) (*model.BindResponse, error) {
	if c.NetworkProfile == "" {
		return nil, errors.New("network profile not configured")
	}
	systemDomain := c.SystemDomain
	providerId := c.ProviderId
	if len(response.Endpoints) == 0 {
		response.Endpoints = response.Credentials.Endpoints
	}
	response.Credentials.Endpoints = nil
	profiles.AddIstioNetworkDataToResponse(providerId, bindingId, systemDomain, c.LoadBalancerPort, &response, c.NetworkProfile)

	err := c.writeIstioFilesForProvider(bindingId, &request, &response)
	if err != nil {
		c.PostDelete(bindingId)
		return nil, err
	}
	return &response, nil
}

func (c ProducerInterceptor) HasAdaptCredentials() bool {
	return true
}

func (c ProducerInterceptor) PostDelete(bindId string) error {
	fileName := path.Join(c.IstioDirectory, bindId) + ".yml"
	err := os.Remove(fileName)
	if err != nil {
		log.Printf("Ignoring error during removal of file %s: %v\n", fileName, err)
	}
	return nil
}

func (c ProducerInterceptor) writeIstioFilesForProvider(bindingId string, request *model.BindRequest, response *model.BindResponse) error {
	return c.writeIstioConfigFiles(bindingId, config.CreateIstioConfigForProvider(request, response, bindingId, c.SystemDomain, c.ProviderId))
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
