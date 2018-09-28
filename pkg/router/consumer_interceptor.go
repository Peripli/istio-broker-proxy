package router

import (
	"fmt"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/config"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/profiles"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	namespace    = "cki75"
	service_port = 5555
)

type Kubernetes interface {
	CreateService(*v1.Service) (*v1.Service, error)
	CreateObject(runtime.Object) error
}

type ConsumerInterceptor struct {
	ConsumerId string
	Kubernetes Kubernetes
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
		c.Kubernetes.CreateService(service)
		configurations := config.CreateEntriesForExternalServiceClient(service.Name, endpoint.Host, "", 0)
		for _, configuration := range configurations {
			runtimeObject, _ := config.ToRuntimeObject(configuration)
			c.Kubernetes.CreateObject(runtimeObject)
		}
	}
	return &response, nil
}

func (c ConsumerInterceptor) postBindExperiment(request model.BindRequest, response model.BindResponse, bindId string) (*model.BindResponse, error) {

	for index, endpoint := range response.NetworkData.Data.Endpoints {
		service := &v1.Service{Spec: v1.ServiceSpec{Ports: []v1.ServicePort{{Port: 5555, TargetPort: intstr.FromInt(5555)}}}}
		service.Name = fmt.Sprintf("service-%d-%s", index, bindId)
		service, err := c.Kubernetes.CreateService(service)
		if err != nil {
			return nil, err
		}
		hostname := endpoint.Host
		configurations := config.CreateEntriesForExternalServiceClient(service.Name, hostname, service.Spec.ClusterIP, endpoint.Port)

		for _, configuration := range configurations {
			runtimeObject, err := config.ToRuntimeObject(configuration)
			if err != nil {
				return nil, err
			}
			c.Kubernetes.CreateObject(runtimeObject)
		}
	}
	return &response, nil
}

func (c ConsumerInterceptor) adaptCredentials(in []byte) ([]byte, error) {
	return in, nil
}

func InClusterServiceFactory() Kubernetes {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return InClusterKubernetes{clientset}

}

type InClusterKubernetes struct {
	*kubernetes.Clientset
}

func (c InClusterKubernetes) CreateService(service *v1.Service) (*v1.Service, error) {
	return service, nil
	//return c.CoreV1().Services(namespace).Create(service)
}

func (c InClusterKubernetes) CreateObject(runtime.Object) error {
	//c.RESTClient().Post().
	//	Namespace(namespace).
	//	Resource(config.Type).
	//	Body(obj).
	//	Do().
	//	Get()

	return nil
}
