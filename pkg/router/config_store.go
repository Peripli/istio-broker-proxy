package router

import (
	"io/ioutil"
	"istio.io/istio/pilot/pkg/config/kube/crd"
	"istio.io/istio/pilot/pkg/model"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

type ConfigStore interface {
	CreateService(*v1.Service) (*v1.Service, error)
	CreateIstioConfig(model.Config) error
}

func NewInClusterConfigStore() ConfigStore {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	namespace, err := getNamespace()
	if err != nil {
		panic(err.Error())
	}
	return newKubeConfigStore(cfg, namespace)
}

func newKubeConfigStore(config *rest.Config, namespace string) ConfigStore {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return kubeConfigStore{clientset, namespace}

}
func getNamespace() (string, error) {
	file, err := os.Open("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return "", err
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

type kubeConfigStore struct {
	*kubernetes.Clientset
	namespace string
}

func (k kubeConfigStore) CreateService(service *v1.Service) (*v1.Service, error) {
	return k.CoreV1().Services(k.namespace).Create(service)
}

func (k kubeConfigStore) CreateIstioConfig(cfg model.Config) error {

	kubeCfgFile := os.Getenv("KUBECONFIG")
	configClient, err := crd.NewClient(kubeCfgFile, "shoot--istio--dev", model.IstioConfigTypes, "cluster.local")
	if err != nil {
		return err
	}

	_, err = configClient.Create(cfg)
	return err
}

func NewExternKubeConfigStore(namespace string) ConfigStore {
	cfg, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		panic(err.Error())
	}
	return newKubeConfigStore(cfg, namespace)

}
