package router

import (
	"fmt"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/config"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/profiles"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"log"
	"strings"
)

const (
	service_port = 5555
)

type ConsumerInterceptor struct {
	ConsumerId      string
	ConfigStore     ConfigStore
	ServiceIdPrefix string
}

func (c ConsumerInterceptor) PreBind(request model.BindRequest) *model.BindRequest {
	request.NetworkData.Data.ConsumerId = c.ConsumerId
	request.NetworkData.NetworkProfileId = profiles.NetworkProfile
	return &request
}

func (c ConsumerInterceptor) PostBind(request model.BindRequest, response model.BindResponse, bindId string,
	adapt func(model.Credentials, []model.EndpointMapping) (*model.BindResponse, error)) (*model.BindResponse, error) {
	var endpointMapping []model.EndpointMapping

	if len(response.NetworkData.Data.Endpoints) != len(response.Endpoints) {
		return nil, fmt.Errorf("Number of endpoints in NetworkData.Data (%d) doesn't match number of endpoints in root (%d)",
			len(response.NetworkData.Data.Endpoints), len(response.Endpoints))
	}
	for index, endpoint := range response.NetworkData.Data.Endpoints {
		clusterIp, err := CreateIstioObjectsInK8S(c.ConfigStore, serviceName(index, bindId), endpoint)
		if err != nil {
			return nil, err
		}
		endpointMapping = append(endpointMapping,
			model.EndpointMapping{
				Source: response.Endpoints[index],
				Target: model.Endpoint{Host: clusterIp, Port: service_port}})
	}
	binding, err := adapt(response.Credentials, endpointMapping)
	if err != nil {
		return nil, err
	}
	binding.NetworkData = response.NetworkData
	binding.AdditionalProperties = response.AdditionalProperties
	return binding, nil
}

func CreateIstioObjectsInK8S(configStore ConfigStore, name string, endpoint model.Endpoint) (string, error) {
	service := &v1.Service{Spec: v1.ServiceSpec{Ports: []v1.ServicePort{{Port: service_port, TargetPort: intstr.FromInt(service_port)}}}}
	service.Name = name
	service, err := configStore.CreateService(service)
	if err != nil {
		log.Println("error creating service:", err.Error())
		return "", err
	}
	configurations := config.CreateEntriesForExternalServiceClient(service.Name, endpoint.Host, service.Spec.ClusterIP, 9000)
	for _, configuration := range configurations {
		err = configStore.CreateIstioConfig(configuration)
		if err != nil {
			log.Printf("error creating %#v: %s\n", configuration, err.Error())
			return "", err
		}
	}
	return service.Spec.ClusterIP, nil
}

func serviceName(index int, bindId string) string {
	name := fmt.Sprintf("svc-%d-%s", index, bindId)
	return name
}

func (c ConsumerInterceptor) PostDelete(bindId string) error {
	i := 0
	var err error

	for err == nil {
		serviceName := serviceName(i, bindId)

		for _, id := range config.DeleteEntriesForExternalServiceClient(serviceName) {
			_ = c.ConfigStore.DeleteIstioConfig(id.Type, id.Name)
		}
		err = c.ConfigStore.DeleteService(serviceName)
		i++
	}
	return nil
}

func (c ConsumerInterceptor) HasAdaptCredentials() bool {
	return false
}

func (c ConsumerInterceptor) PostCatalog(catalog *model.Catalog) {
	for i := range catalog.Services {
		catalog.Services[i].Id = strings.TrimPrefix(catalog.Services[i].Id, c.ServiceIdPrefix)
	}
}
