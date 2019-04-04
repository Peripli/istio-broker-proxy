package router

import (
	"errors"
	"fmt"
	istioModel "istio.io/istio/pilot/pkg/model"
	"k8s.io/api/core/v1"
)

type mockConfigStore struct {
	CreatedServices      []*v1.Service
	CreatedIstioConfigs  []istioModel.Config
	ClusterIP            string
	CreateServiceErr     error
	CreateObjectErr      error
	CreateObjectErrCount int
	DeletedServices      []string
	DeletedIstioConfigs  []string
}

func (m *mockConfigStore) CreateService(bindingID string, service *v1.Service) (*v1.Service, error) {
	if m.CreateServiceErr != nil {
		return nil, m.CreateServiceErr
	}
	if service.Labels == nil {
		service.Labels = make(map[string]string)
	}
	service.Labels["istio-broker-proxy-binding-id"] = bindingID
	m.CreatedServices = append(m.CreatedServices, service)
	service.Spec.ClusterIP = m.ClusterIP
	return service, nil
}

func (m *mockConfigStore) getNamespace() string {
	return "catalog"
}

func (m *mockConfigStore) CreateIstioConfig(bindingID string, configs []istioModel.Config) error {
	for _, config := range configs {
		if config.Labels == nil {
			config.Labels = make(map[string]string)
		}
		config.Labels["istio-broker-proxy-binding-id"] = bindingID
		if m.CreateObjectErr != nil && m.CreateObjectErrCount == len(m.CreatedIstioConfigs) {
			return m.CreateObjectErr
		}
		m.CreatedIstioConfigs = append(m.CreatedIstioConfigs, config)
	}
	return nil
}

func (m *mockConfigStore) deleteService(bindingID string) error {
	found := 0
	services := append([]*v1.Service{}, m.CreatedServices...)

	for index, c := range services {
		if c.Labels["istio-broker-proxy-binding-id"] == bindingID {
			m.DeletedServices = append(m.DeletedServices, c.Name)
			m.CreatedServices = append(m.CreatedServices[:index-found], m.CreatedServices[index-found+1:]...)
			found++
		}
	}
	if found == 0 {
		errorMsg := fmt.Sprintf("error binding-id %s not found", bindingID)
		return errors.New(errorMsg)
	}
	return nil
}

func (m *mockConfigStore) DeleteBinding(bindingID string) error {
	found := 0
	configs := append([]istioModel.Config{}, m.CreatedIstioConfigs...)
	for index, c := range configs {
		if c.Labels["istio-broker-proxy-binding-id"] == bindingID {
			m.DeletedIstioConfigs = append(m.DeletedIstioConfigs, c.Type+":"+c.Name)
			m.CreatedIstioConfigs = append(m.CreatedIstioConfigs[:index-found], m.CreatedIstioConfigs[index-found+1:]...)
			found++
		}
	}
	if found == 0 {
		errorMsg := fmt.Sprintf("error binding-id %s not found", bindingID)
		return errors.New(errorMsg)
	}
	return m.deleteService(bindingID)
}

//NewMockConfigStore create a new ConfigStore with mocking capabilities
func NewMockConfigStore() ConfigStore {
	return &mockConfigStore{}
}
