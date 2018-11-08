package router

import (
	"fmt"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/config"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/profiles"
	istioModel "istio.io/istio/pilot/pkg/model"
	"log"
	"os"
	"path"
)

type ProducerInterceptor struct {
	LoadBalancerPort int
	SystemDomain     string
	ProviderId       string
	IstioDirectory   string
	IpAddress        string
	ServiceIdPrefix  string
}

func (c *ProducerInterceptor) WriteIstioConfigFiles(port int) error {
	return c.writeIstioConfigFiles("istio-broker",
		config.CreateEntriesForExternalService("istio-broker", string(c.IpAddress), uint32(port), "istio-broker."+c.SystemDomain, "client.istio.sapcloud.io", 9000))
}

func (c ProducerInterceptor) PreBind(request model.BindRequest) *model.BindRequest {
	return &request
}

func (c ProducerInterceptor) PostBind(request model.BindRequest, response model.BindResponse, bindingId string,
	adapt func(model.Credentials, []model.EndpointMapping) (*model.BindResponse, error)) (*model.BindResponse, error) {
	systemDomain := c.SystemDomain
	providerId := c.ProviderId
	if len(response.Endpoints) == 0 {
		response.Endpoints = response.Credentials.Endpoints
	}
	response.Credentials.Endpoints = nil
	profiles.AddIstioNetworkDataToResponse(providerId, bindingId, systemDomain, c.LoadBalancerPort, &response)

	err := c.writeIstioFilesForProvider(bindingId, &request, &response)
	if err != nil {
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
		fmt.Printf("Ignoring error during removal of file %s: %v\n", fileName, err)
	}
	return nil
}

func (c ProducerInterceptor) writeIstioFilesForProvider(bindingId string, request *model.BindRequest, response *model.BindResponse) error {
	return c.writeIstioConfigFiles(bindingId, config.CreateIstioConfigForProvider(request, response, bindingId, c.SystemDomain))
}

func (c ProducerInterceptor) writeIstioConfigFiles(fileName string, configuration []istioModel.Config) error {
	ymlPath := path.Join(c.IstioDirectory, fileName) + ".yml"
	log.Printf("PATH to istio config: %v\n", ymlPath)
	file, err := os.Create(ymlPath)
	if nil != err {
		return err
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

func (c ProducerInterceptor) PostCatalog(catalog *model.Catalog) {
	for i := range catalog.Services {
		catalog.Services[i].Id = c.ServiceIdPrefix + catalog.Services[i].Id
	}
}
