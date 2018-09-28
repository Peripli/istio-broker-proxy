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

type ProducerConfig struct {
	LoadBalancerPort int
	SystemDomain     string
	ProviderId       string
	IstioDirectory   string
	IpAddress        string
}

type producer_interceptor struct {
	config ProducerConfig
}

func NewProducerInterceptor(cfg ProducerConfig, port int) ServiceBrokerInterceptor {
	if cfg.IpAddress == "" {
		cfg.IpAddress = "127.0.0.1"
	}
	if cfg.IstioDirectory == "" {
		cfg.IstioDirectory = os.TempDir()
	}
	if cfg.LoadBalancerPort == 0 {
		cfg.LoadBalancerPort = 9000
	}
	interceptor := producer_interceptor{cfg}
	interceptor.writeIstioConfigFiles("istio-broker",
		config.CreateEntriesForExternalService("istio-broker", string(cfg.IpAddress), uint32(port), "istio-broker."+cfg.SystemDomain, "client.istio.sapcloud.io", 9000))
	return &interceptor
}

func (c producer_interceptor) preBind(request model.BindRequest) *model.BindRequest {
	return &request
}

func (c producer_interceptor) postBind(request model.BindRequest, response model.BindResponse, bindingId string) (*model.BindResponse, error) {
	systemDomain := c.config.SystemDomain
	providerId := c.config.ProviderId
	if len(response.Endpoints) == 0 {
		response.Endpoints = response.Credentials.Endpoints
	}
	profiles.AddIstioNetworkDataToResponse(providerId, bindingId, systemDomain, c.config.LoadBalancerPort, &response)

	err := c.writeIstioFilesForProvider(bindingId, &request, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (c producer_interceptor) writeIstioFilesForProvider(bindingId string, request *model.BindRequest, response *model.BindResponse) error {
	return c.writeIstioConfigFiles(bindingId, config.CreateIstioConfigForProvider(request, response, bindingId, c.config.SystemDomain))
}

func (c producer_interceptor) writeIstioConfigFiles(fileName string, configuration []istioModel.Config) error {
	ymlPath := path.Join(c.config.IstioDirectory, fileName) + ".yml"
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
