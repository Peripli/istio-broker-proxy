package main

import (
	"github.com/Peripli/service-manager/pkg/web"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/plugin"
	"unsafe"
)

func Init(api unsafe.Pointer) error {
	myApi := ((*web.API)(api))
	plugin.InitSimplePlugin(myApi)
	return nil
}
