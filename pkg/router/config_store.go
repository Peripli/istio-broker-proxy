package router

import (
	"istio.io/istio/pilot/pkg/model"
	"k8s.io/api/core/v1"
)

//ConfigStore encapsulates functions to modify istio config
type ConfigStore interface {
	CreateService(bindingID string, service *v1.Service) (*v1.Service, error)
	CreateIstioConfig(bindingID string, config []model.Config) error
	DeleteBinding(bindingID string) error
}
