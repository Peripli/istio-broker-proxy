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

func (c ConsumerInterceptor) postBind(request model.BindRequest, response model.BindResponse, bindId string) (*model.BindResponse, error) {
	for index, endpoint := range response.NetworkData.Data.Endpoints {
		service := &v1.Service{Spec: v1.ServiceSpec{Ports: []v1.ServicePort{{Port: service_port, TargetPort: intstr.FromInt(service_port)}}}}
		name := fmt.Sprintf("svc-%d-%s", index, bindId)
		service.Name = name
		service, err := c.ConfigStore.CreateService(service)
		if err != nil {
			log.Println("error creating service")
			return nil, err
		}
		configurations := config.CreateEntriesForExternalServiceClient(service.Name, endpoint.Host, service.Spec.ClusterIP, 17171)
		for _, configuration := range configurations {
			err = c.ConfigStore.CreateIstioConfig(configuration)
			if err != nil {
				log.Printf("error creating %#v\n", configuration)
				return nil, err
			}
		}
	}
	return &response, nil
}

func (c ConsumerInterceptor) hasAdaptCredentials() bool {
	return false
}
