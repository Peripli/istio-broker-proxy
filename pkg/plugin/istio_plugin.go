package plugin

import (
	"encoding/json"
	"github.com/Peripli/service-manager/pkg/web"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/router"
	"strings"
)

type ItioPlugin struct {
	interceptor router.ServiceBrokerInterceptor
}

func (i *ItioPlugin) Name() string {
	return "istio"
}

func (i *ItioPlugin) Bind(request *web.Request, next web.Handler) (*web.Response, error) {
	var bindRequest model.BindRequest
	json.Unmarshal(request.Body, &bindRequest)
	bindRequest = *i.interceptor.PreBind(bindRequest)
	request.Body, _ = json.Marshal(bindRequest)
	response, _ := next.Handle(request)
	var bindResponse model.BindResponse
	json.Unmarshal(response.Body, &bindResponse)
	modifiedBindResponse, _ := i.interceptor.PostBind(bindRequest, bindResponse, extractBindId(request.URL.Path), model.Adapt)
	response.Body, _ = json.Marshal(modifiedBindResponse)
	return response, nil
}

func (i *ItioPlugin) Unbind(request *web.Request, next web.Handler) (*web.Response, error) {
	// call interceptor.PostDelete()
	return next.Handle(request)
}

func extractBindId(path string) string {
	return strings.Split(path, "/")[3]
}

func InitItioPlugin(api *web.API) {
	istioPlugin := &ItioPlugin{interceptor: &router.NoOpInterceptor{}}
	api.RegisterPlugins(istioPlugin)
}
