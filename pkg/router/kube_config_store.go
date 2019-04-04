package router

import (
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

//NewInClusterConfigStore creates a new ConfigStore from within the cluster
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

func (k kubeConfigStore) getNamespace() string {
	return k.namespace
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

func (k kubeConfigStore) CreateService(bindingID string, service *v1.Service) (*v1.Service, error) {
	service.Labels["istio-broker-proxy-binding-id"] = bindingID
	return k.CoreV1().Services(k.namespace).Create(service)
}

func (k kubeConfigStore) CreateIstioConfig(bindingID string, configurations []model.Config) error {
	for _, config := range configurations {
		config.Labels["istio-broker-proxy-binding-id"] = bindingID
		_, err := k.configClient.Create(config)
		if err != nil {
			log.Printf("error creating %s: %s\n", config.Name, err.Error())
			return err
		}
	}
	return nil
}

func (k kubeConfigStore) DeleteService(serviceName string) error {
	log.Printf("kubectl -n %s delete services %s\n", k.namespace, serviceName)
	err := k.CoreV1().Services(k.namespace).Delete(serviceName, &meta_v1.DeleteOptions{})
	if err != nil {
		log.Printf("error %s\n", err.Error())
	}
	return err
}

func (k kubeConfigStore) DeleteIstioConfig(configType string, configName string) error {
	log.Printf("kubectl -n %s delete %s %s\n", k.namespace, configType, configName)
	err := k.configClient.Delete(configType, configName, k.namespace)
	if err != nil {
		log.Printf("error %s\n", err.Error())
	}
	return err
}

//NewExternKubeConfigStore creates a new ConfigStore using the KUBECONFIG env variable
func NewExternKubeConfigStore(namespace string) ConfigStore {
	clientcmd.ClusterDefaults.Server = ""
	cfg, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		panic(err.Error())
	}
	return newKubeConfigStore(cfg, namespace)

}
