package plugin

import (
	"github.com/Peripli/service-manager/pkg/web"
)

type SimplePlugin struct {
}

func (i *SimplePlugin) Name() string {
	return "simple"
}

func (i *SimplePlugin) Bind(request *web.Request, next web.Handler) (*web.Response, error) {
	request.Header.Add("Hello", "World")
	return next.Handle(request)
}

func InitSimplePlugin(api *web.API) error {
	simplePlugin := &SimplePlugin{}
	api.RegisterPlugins(simplePlugin)
	return nil
}
