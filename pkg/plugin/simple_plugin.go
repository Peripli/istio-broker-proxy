package plugin

import (
	"github.com/Peripli/service-manager/pkg/web"
	"log"
)

type SimplePlugin struct {
}

func (i *SimplePlugin) Name() string {
	return "simple"
}

func (i *SimplePlugin) Bind(request *web.Request, next web.Handler) (*web.Response, error) {
	request.Header.Add("Hello", "World")
	response, err := next.Handle(request)
	log.Printf("SimplePlugin was triggered with request body: %s\n", string(request.Body))
	return response, err
}

func InitSimplePlugin(api *web.API) error {
	simplePlugin := &SimplePlugin{}
	api.RegisterPlugins(simplePlugin)
	return nil
}
