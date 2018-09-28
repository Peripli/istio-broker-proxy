package router

import (
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/profiles"
)

type ConsumerInterceptor struct {
	ConsumerId string
}

func (c ConsumerInterceptor) preBind(request model.BindRequest) *model.BindRequest {
	request.NetworkData.Data.ConsumerId = c.ConsumerId
	request.NetworkData.NetworkProfileId = profiles.NetworkProfile
	return &request
}

func (c ConsumerInterceptor) postBind(request model.BindRequest, response model.BindResponse, bindId string) (*model.BindResponse, error) {
	return &response, nil
}

func (c ConsumerInterceptor) adaptCredentials(in []byte) ([]byte, error) {
	return in, nil
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
