package router

import (
	"context"
	"github.com/ericchiang/k8s"
	xv1 "github.com/ericchiang/k8s/apis/core/v1"
	v12 "github.com/ericchiang/k8s/apis/meta/v1"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"istio.io/istio/pilot/pkg/model"
	"k8s.io/api/core/v1"
	"os"
)

func NewInClusterEricConfigStore() ConfigStore {
	namespace, err := getNamespace()
	if err != nil {
		panic(err.Error())
	}
	return newEricKubeConfigStore(nil, namespace)
}

func newEricKubeConfigStore(k8sConfig *k8s.Config, namespace string) ConfigStore {

	k8s.Register("networking.istio.io", "v1alpha3", "serviceentries", true, &serviceEntryResource{})

	client, err := k8s.NewClient(k8sConfig)
	if err != nil {
		panic(err.Error())
	}

	return ericKubeConfigStore{client, namespace}

}

type ericKubeConfigStore struct {
	client    *k8s.Client
	namespace string
}

func (k ericKubeConfigStore) CreateService(service *v1.Service) (*v1.Service, error) {
	var x xv1.Service
	y, err := yaml.Marshal(service)
	if err != nil {
		return nil, err
	}
	yaml.Unmarshal(y, &x)
	x.Metadata.Namespace = &k.namespace
	err = k.client.Create(context.TODO(), &x)
	if err != nil {
		return nil, err
	}
	k.client.Get(context.TODO(), k.namespace, service.Name, &x)
	y, err = yaml.Marshal(x)
	if err != nil {
		return nil, err
	}
	yaml.Unmarshal(y, &service)
	return service, nil
}

func (k ericKubeConfigStore) CreateIstioConfig(cfg model.Config) error {

	resource, err := k.convertConfig(cfg)
	if err != nil {
		return err
	}
	err = k.client.Create(context.TODO(), resource)

	return err
}

func (k ericKubeConfigStore) convertConfig(config model.Config) (k8s.Resource, error) {
	spec, err := model.ToJSONMap(config.Spec)
	if err != nil {
		return nil, err
	}
	namespace := config.Namespace
	if namespace == "" {
		namespace = k.namespace
	}
	var out k8s.Resource
	out = &serviceEntryResource{
		Kind:       "ServiceEntry",
		ApiVersion: "networking.istio.io/v1alpha3",
		MetaData: v12.ObjectMeta{
			Name:            &config.Name,
			Namespace:       &namespace,
			ResourceVersion: &config.ResourceVersion,
			Labels:          config.Labels,
			Annotations:     config.Annotations,
		},
		Spec: spec,
	}
	return out, nil
}

type serviceEntryResource struct {
	ApiVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	MetaData   v12.ObjectMeta         `json:"metadata"`
	Spec       map[string]interface{} `json:"spec"`
}

func (s serviceEntryResource) GetMetadata() *v12.ObjectMeta {
	return &s.MetaData
}

func NewExternEricKubeConfigStore(namespace string) ConfigStore {
	data, err := ioutil.ReadFile(os.Getenv("KUBECONFIG"))
	if err != nil {
		panic(err.Error())
	}

	var k8sConfig k8s.Config
	if err := yaml.Unmarshal(data, &k8sConfig); err != nil {
		panic(err.Error())
	}
	return newEricKubeConfigStore(&k8sConfig, namespace)

}
