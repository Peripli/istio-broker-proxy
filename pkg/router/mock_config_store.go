package router

import (
	"errors"
	"fmt"
	istioModel "istio.io/istio/pilot/pkg/model"
	"k8s.io/api/core/v1"
)

type MockConfigStore struct {
	CreatedServices      []*v1.Service
	CreatedIstioConfigs  []istioModel.Config
	ClusterIp            string
	CreateServiceErr     error
	CreateObjectErr      error
	CreateObjectErrCount int
	DeletedServices      []string
	DeletedIstioConfigs  []string
}

func (m *MockConfigStore) CreateService(service *v1.Service) (*v1.Service, error) {
	if m.CreateServiceErr != nil {
		return nil, m.CreateServiceErr
	}
	m.CreatedServices = append(m.CreatedServices, service)
	service.Spec.ClusterIP = m.ClusterIp
	return service, nil
}

func (m *MockConfigStore) getNamespace() string {
	return "catalog"
}

func (m *MockConfigStore) CreateIstioConfig(object istioModel.Config) error {
	if m.CreateObjectErr != nil && m.CreateObjectErrCount == len(m.CreatedIstioConfigs) {
		return m.CreateObjectErr
	}
	m.CreatedIstioConfigs = append(m.CreatedIstioConfigs, object)
	return nil
}

func (m *MockConfigStore) DeleteService(serviceName string) error {
	for index, c := range m.CreatedServices {
		if c.Name == serviceName {
			m.DeletedServices = append(m.DeletedServices, serviceName)
			m.CreatedServices = append(m.CreatedServices[:index], m.CreatedServices[index+1:]...)
			return nil
		}
	}
	errorMsg := fmt.Sprintf("error services %s not found", serviceName)
	return errors.New(errorMsg)
}

func (m *MockConfigStore) DeleteIstioConfig(configType string, configName string) error {
	for index, c := range m.CreatedIstioConfigs {
		if c.Name == configName {
			m.DeletedIstioConfigs = append(m.DeletedIstioConfigs, configType+":"+configName)
			m.CreatedIstioConfigs = append(m.CreatedIstioConfigs[:index], m.CreatedIstioConfigs[index+1:]...)
			return nil
		}
	}
	errorMsg := fmt.Sprintf("error %s.networking.istio.io %s not found", configType, configName)
	return errors.New(errorMsg)
}
