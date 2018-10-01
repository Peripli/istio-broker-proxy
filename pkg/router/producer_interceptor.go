package router

import (
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
}

func (c *ProducerInterceptor) WriteIstioConfigFiles(port int) {
	c.writeIstioConfigFiles("istio-broker",
		config.CreateEntriesForExternalService("istio-broker", string(c.IpAddress), uint32(port), "istio-broker."+c.SystemDomain, "client.istio.sapcloud.io", 9000))
}

func (c ProducerInterceptor) preBind(request model.BindRequest) *model.BindRequest {
	return &request
}

func (c ProducerInterceptor) postBind(request model.BindRequest, response model.BindResponse, bindingId string) (*model.BindResponse, error) {
	systemDomain := c.SystemDomain
	providerId := c.ProviderId
	if len(response.Endpoints) == 0 {
		response.Endpoints = response.Credentials.Endpoints
	}
	profiles.AddIstioNetworkDataToResponse(providerId, bindingId, systemDomain, c.LoadBalancerPort, &response)

	err := c.writeIstioFilesForProvider(bindingId, &request, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (c ProducerInterceptor) hasAdaptCredentials() bool {
	return true
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
