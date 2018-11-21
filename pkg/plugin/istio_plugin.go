package plugin

import (
	"encoding/json"
	"github.com/Peripli/service-manager/pkg/web"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/router"
	"log"
	"strings"
)

type IstioPlugin struct {
	interceptor router.ServiceBrokerInterceptor
}

func (i *IstioPlugin) Name() string {
	return "istio"
}

func (i *IstioPlugin) Bind(request *web.Request, next web.Handler) (*web.Response, error) {
	var bindRequest model.BindRequest
	log.Printf("IstioPlugin bind was triggered with request body: %s\n", string(request.Body))
	json.Unmarshal(request.Body, &bindRequest)
	log.Println("execute prebind")
	bindRequest = *i.interceptor.PreBind(bindRequest)
	request.Body, _ = json.Marshal(bindRequest)
	response, _ := next.Handle(request)
	var bindResponse model.BindResponse
	json.Unmarshal(response.Body, &bindResponse)
	log.Println("execute postbind")
	modifiedBindResponse, err := i.interceptor.PostBind(bindRequest, bindResponse, extractBindId(request.URL.Path), model.Adapt)
	if err != nil {
		log.Printf("Error during PostBind %s\n", err.Error())
	}
	response.Body, _ = json.Marshal(modifiedBindResponse)
	return response, nil
}

func (i *IstioPlugin) Unbind(request *web.Request, next web.Handler) (*web.Response, error) {
	log.Printf("IstioPlugin unbind was triggered with request body: %s\n", string(request.Body))
	i.interceptor.PostDelete(extractBindId(request.URL.Path))
	return next.Handle(request)
}

func extractBindId(path string) string {
	return strings.Split(path, "/")[3]
}

func createConsumerInterceptor() router.ConsumerInterceptor {
	consumerInterceptor := router.ConsumerInterceptor{}
	consumerInterceptor.ServiceIdPrefix = "istio-"
	consumerInterceptor.ConsumerId = "client.istio.sapcloud.io"
	consumerInterceptor.ConfigStore = router.NewInClusterConfigStore()
	return consumerInterceptor
}

func InitIstioPlugin(api *web.API) {
	istioPlugin := &IstioPlugin{interceptor: createConsumerInterceptor()}
	api.RegisterPlugins(istioPlugin)
}
