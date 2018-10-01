package router

import (
	"fmt"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/config"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/profiles"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	service_port = 5555
)

type ConsumerInterceptor struct {
	ConsumerId  string
	ConfigStore ConfigStore
}

func (c ConsumerInterceptor) preBind(request model.BindRequest) *model.BindRequest {
	request.NetworkData.Data.ConsumerId = c.ConsumerId
	request.NetworkData.NetworkProfileId = profiles.NetworkProfile
	return &request
}

func (c ConsumerInterceptor) postBind(request model.BindRequest, response model.BindResponse, bindId string) (*model.BindResponse, error) {
	for index, endpoint := range response.NetworkData.Data.Endpoints {
		service := &v1.Service{Spec: v1.ServiceSpec{Ports: []v1.ServicePort{{Port: service_port, TargetPort: intstr.FromInt(service_port)}}}}
		service.Name = fmt.Sprintf("service-%d-%s", index, bindId)
		service, err := c.ConfigStore.CreateService(service)
		if err != nil {
			return nil, err
		}
		configurations := config.CreateEntriesForExternalServiceClient(service.Name, endpoint.Host, service.Spec.ClusterIP, 0)
		for _, configuration := range configurations {
			err = c.ConfigStore.CreateIstioConfig(configuration)
			if err != nil {
				return nil, err
			}
		}
	}
	return &response, nil
}

func (c ConsumerInterceptor) postBindExperiment(request model.BindRequest, response model.BindResponse, bindId string) (*model.BindResponse, error) {

	for index, endpoint := range response.NetworkData.Data.Endpoints {
		service := &v1.Service{Spec: v1.ServiceSpec{Ports: []v1.ServicePort{{Port: 5555, TargetPort: intstr.FromInt(5555)}}}}
		service.Name = fmt.Sprintf("service-%d-%s", index, bindId)
		service, err := c.ConfigStore.CreateService(service)
		if err != nil {
			return nil, err
		}
		hostname := endpoint.Host
		configurations := config.CreateEntriesForExternalServiceClient(service.Name, hostname, service.Spec.ClusterIP, endpoint.Port)

		for _, configuration := range configurations {
			c.ConfigStore.CreateIstioConfig(configuration)
		}
	}
	return &response, nil
}

func (c ConsumerInterceptor) hasAdaptCredentials() bool {
	return false
}
