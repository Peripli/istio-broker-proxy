package router

import (
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/profiles"
)

type ConsumerConfig struct {
	ConsumerId string
}

type consumer_interceptor struct {
	config ConsumerConfig
}

func NewConsumerInterceptor(cfg ConsumerConfig) ServiceBrokerInterceptor {
	return &consumer_interceptor{cfg}
}

func (c consumer_interceptor) preBind(request model.BindRequest) *model.BindRequest {
	request.NetworkData.Data.ConsumerId = c.config.ConsumerId
	request.NetworkData.NetworkProfileId = profiles.NetworkProfile
	return &request
}

func (c consumer_interceptor) postBind(request model.BindRequest, response model.BindResponse, bindId string) (*model.BindResponse, error) {
	return &response, nil
}

//func inClusterServiceFactory() clientv1.ServiceInterface {
//	config, err := rest.InClusterConfig()
//	if err != nil {
//		panic(err.Error())
//	}
//	clientset, err := kubernetes.NewForConfig(config)
//	if err != nil {
//		panic(err.Error())
//	}
//	return clientset.CoreV1().Services("istio-system")
//
//}
//
//func test() {
//	inClusterServiceFactory().Create(&v1.Service{})
//}
