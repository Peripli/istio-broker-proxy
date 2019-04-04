package router

import (
	"errors"
	"fmt"
	"github.com/Peripli/istio-broker-proxy/pkg/config"
	"github.com/Peripli/istio-broker-proxy/pkg/model"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"log"
	"strings"
)

const (
	servicePort = 5555
)

//ConsumerInterceptor contains config for the consumer side
type ConsumerInterceptor struct {
	ConsumerID        string
	ConfigStore       ConfigStore
	ServiceNamePrefix string
	NetworkProfile    string
}

//PreBind see interface definition
func (c ConsumerInterceptor) PreBind(request model.BindRequest) (*model.BindRequest, error) {
	if c.NetworkProfile == "" {
		return nil, errors.New("network profile not configured")
	}
	request.NetworkData.Data.ConsumerID = c.ConsumerID
	request.NetworkData.NetworkProfileID = c.NetworkProfile
	return &request, nil
}

//PostBind see interface definition
func (c ConsumerInterceptor) PostBind(request model.BindRequest, response model.BindResponse, bindID string,
	adapt func(model.Credentials, []model.EndpointMapping) (*model.BindResponse, error)) (*model.BindResponse, error) {
	var endpointMapping []model.EndpointMapping

	networkDataMatches := (c.NetworkProfile == response.NetworkData.NetworkProfileID)
	if !networkDataMatches {
		log.Println("Ignoring bind request for network id:", response.NetworkData.NetworkProfileID)
		return &response, nil
	}

	if len(response.NetworkData.Data.Endpoints) != len(response.Endpoints) {
		return nil, fmt.Errorf("Number of endpoints in NetworkData.Data (%d) doesn't match number of endpoints in root (%d)",
			len(response.NetworkData.Data.Endpoints), len(response.Endpoints))
	}

	endCleanupCondition := func(index int, err error) bool {
		return index >= len(response.NetworkData.Data.Endpoints)
	}

	log.Printf("Number of endpoints: %d\n", len(response.NetworkData.Data.Endpoints))
	for index, endpoint := range response.NetworkData.Data.Endpoints {
		clusterIP, err := CreateIstioObjectsInK8S(c.ConfigStore, bindID, serviceName(index, bindID), endpoint, response.NetworkData.Data.ProviderID)
		if err != nil {
			c.cleanUpConfig(bindID, endCleanupCondition)
			return nil, err
		}
		endpointMapping = append(endpointMapping,
			model.EndpointMapping{
				Source: response.Endpoints[index],
				Target: model.Endpoint{Host: clusterIP, Port: servicePort}})
	}
	binding, err := adapt(response.Credentials, endpointMapping)
	if err != nil {
		c.cleanUpConfig(bindID, endCleanupCondition)
		return nil, err
	}
	binding.NetworkData = response.NetworkData
	binding.AdditionalProperties = response.AdditionalProperties
	return binding, nil
}

//CreateIstioObjectsInK8S create a service and istio routing rules
func CreateIstioObjectsInK8S(configStore ConfigStore, bindingID string, name string, endpoint model.Endpoint, systemDomain string) (string, error) {
	service := &v1.Service{Spec: v1.ServiceSpec{Ports: []v1.ServicePort{{Port: servicePort, TargetPort: intstr.FromInt(servicePort)}}}}
	service.Name = name
	log.Println("Creating istio objects for", name)
	service, err := configStore.CreateService(bindingID, service)
	if err != nil {
		log.Println("error creating service:", err.Error())
		return "", err
	}
	configurations := config.CreateEntriesForExternalServiceClient(service.Name, endpoint.Host, service.Spec.ClusterIP, 9000, configStore.getNamespace(), systemDomain)
	err = configStore.CreateIstioConfig(bindingID, configurations)
	if err != nil {
		return "", err
	}
	return service.Spec.ClusterIP, nil
}

func serviceName(index int, bindID string) string {
	name := fmt.Sprintf("svc-%d-%s", index, bindID)
	return name
}

//PostDelete see interface definition
func (c ConsumerInterceptor) PostDelete(bindID string) {
	c.cleanUpConfig(bindID, func(index int, err error) bool {
		return err != nil && index > 2
	})
}

func (c ConsumerInterceptor) cleanUpConfig(bindID string, endCleanupCondition func(index int, err error) bool) {
	i := 0
	var err error

	for {
		serviceName := serviceName(i, bindID)
		isFirstIteration := i == 0

		for _, id := range config.DeleteEntriesForExternalServiceClient(serviceName) {
			ignoredErr := c.ConfigStore.DeleteIstioConfig(id.Type, id.Name)
			if ignoredErr != nil && isFirstIteration {
				log.Printf("Ignoring error during removal of configuration %s: %s\n", id, ignoredErr.Error())
			}
		}
		err = c.ConfigStore.DeleteService(serviceName)
		if endCleanupCondition(i, err) {
			break
		}
		if err != nil && isFirstIteration {
			log.Printf("Ignoring error during removal of configuration %s: %s\n", serviceName, err.Error())
		}
		i++
	}
}

//HasAdaptCredentials see interface definition
func (c ConsumerInterceptor) HasAdaptCredentials() bool {
	return false
}

//PostCatalog see interface definition
func (c ConsumerInterceptor) PostCatalog(catalog *model.Catalog) error {
	for i := range catalog.Services {
		catalog.Services[i].Name = strings.TrimPrefix(catalog.Services[i].Name, c.ServiceNamePrefix)
	}
	return nil
}
