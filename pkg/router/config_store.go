package router

import (
	"fmt"
	"io/ioutil"
	"istio.io/istio/pilot/pkg/config/kube/crd"
	"istio.io/istio/pilot/pkg/model"
	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
)

type ConfigStore interface {
	CreateService(*v1.Service) (*v1.Service, error)
	CreateIstioConfig(model.Config) error
	DeleteService(string) error
	DeleteIstioConfig(string, string) error
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
	kubeCfgFile := os.Getenv("KUBECONFIG")
	configClient, err := crd.NewClient(kubeCfgFile, "", model.IstioConfigTypes, "cluster.local")
	if err != nil {
		panic(err.Error())
	}

	return kubeConfigStore{clientset, namespace, configClient}

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
	namespace := string(content)
	log.Println("Using namespace", namespace)
	return namespace, nil
}

type kubeConfigStore struct {
	*kubernetes.Clientset
	namespace    string
	configClient *crd.Client
}

func (k kubeConfigStore) CreateService(service *v1.Service) (*v1.Service, error) {
	return k.CoreV1().Services(k.namespace).Create(service)
}

func (k kubeConfigStore) CreateIstioConfig(cfg model.Config) error {
	_, err := k.configClient.Create(cfg)
	return err
}

func (k kubeConfigStore) DeleteService(serviceName string) error {
	fmt.Printf("kubectl -n %s delete services %s\n", k.namespace, serviceName)
	err := k.CoreV1().Services(k.namespace).Delete(serviceName, &meta_v1.DeleteOptions{})
	if err != nil {
		fmt.Printf("error %s\n", err.Error())
	}
	return err
}

func (k kubeConfigStore) DeleteIstioConfig(configType string, configName string) error {
	fmt.Printf("kubectl -n %s delete %s %s\n", k.namespace, configType, configName)
	err := k.configClient.Delete(configType, configName, k.namespace)
	if err != nil {
		fmt.Printf("error %s\n", err.Error())
	}
	return err
}

func NewExternKubeConfigStore(namespace string) ConfigStore {
	clientcmd.ClusterDefaults.Server = ""
	cfg, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		panic(err.Error())
	}
	return newKubeConfigStore(cfg, namespace)

}
