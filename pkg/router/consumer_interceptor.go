package router

import (
	"fmt"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/config"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/profiles"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"log"
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

func (c ConsumerInterceptor) postBind(request model.BindRequest, response model.BindResponse, bindId string,
	adapt func(model.Credentials, []model.EndpointMapping) (*model.BindResponse, error)) (*model.BindResponse, error) {
	var endpointMapping []model.EndpointMapping

	for index, endpoint := range response.Endpoints {
		service := &v1.Service{Spec: v1.ServiceSpec{Ports: []v1.ServicePort{{Port: service_port, TargetPort: intstr.FromInt(service_port)}}}}
		name := c.serviceName(index, bindId)
		service.Name = name
		service, err := c.ConfigStore.CreateService(service)
		if err != nil {
			log.Println("error creating service")
			return nil, err
		}
		configurations := config.CreateEntriesForExternalServiceClient(service.Name, endpoint.Host, service.Spec.ClusterIP, 9000)
		for _, configuration := range configurations {
			err = c.ConfigStore.CreateIstioConfig(configuration)
			if err != nil {
				log.Printf("error creating %#v\n", configuration)
				return nil, err
			}
		}
		endpointMapping = append(endpointMapping,
			model.EndpointMapping{
				Source: endpoint,
				Target: model.Endpoint{Host: service.Spec.ClusterIP, Port: service_port}})
	}
	binding, err := adapt(response.Credentials, endpointMapping)
	if err != nil {
		return nil, err
	}
	binding.NetworkData = response.NetworkData
	binding.AdditionalProperties = response.AdditionalProperties
	return binding, nil
}

func (c ConsumerInterceptor) serviceName(index int, bindId string) string {
	name := fmt.Sprintf("svc-%d-%s", index, bindId)
	return name
}

func (c ConsumerInterceptor) postDelete(bindId string) error {

	serviceName := c.serviceName(0, bindId)
	for _, id := range config.DeleteEntriesForExternalServiceClient(serviceName) {
		_ = c.ConfigStore.DeleteIstioConfig(id.Type, id.Name)
	}
	_ = c.ConfigStore.DeleteService(serviceName)
	return nil
}

func (c ConsumerInterceptor) hasAdaptCredentials() bool {
	return false
}
