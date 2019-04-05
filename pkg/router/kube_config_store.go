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
	"strings"
)

const bindingIDLabel = "istio-broker-proxy-binding-id"

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
	if service.Labels == nil {
		service.Labels = make(map[string]string)
	}
	service.Namespace = k.namespace
	service.Labels[bindingIDLabel] = bindingID
	return k.CoreV1().Services(k.namespace).Create(service)
}

func (k kubeConfigStore) CreateIstioConfig(bindingID string, configurations []model.Config) error {
	for _, config := range configurations {
		if config.Labels == nil {
			config.Labels = make(map[string]string)
		}
		config.Namespace = k.namespace
		config.Labels[bindingIDLabel] = bindingID
		_, err := k.configClient.Create(config)
		if err != nil {
			log.Printf("error creating %s: %s\n", config.Name, err.Error())
			return err
		}
	}
	return nil
}

func (k kubeConfigStore) DeleteBinding(bindingID string) error {
	log.Printf("kubectl -n %s delete services -l %s=%s\n", k.namespace, bindingIDLabel, bindingID)
	services := k.CoreV1().Services(k.namespace)
	list, err := services.List(meta_v1.ListOptions{LabelSelector: bindingIDLabel + "=" + bindingID})
	if err != nil {
		return err
	}
	for _, service := range list.Items {
		err := k.CoreV1().Services(k.namespace).Delete(service.Name, &meta_v1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	for _, typ := range []string{"gateway", "virtual-service", "destination-rule", "service-entry"} {
		log.Printf("kubectl -n %s delete %s -l %s=%s\n", k.namespace, strings.Replace(typ,"-","",-1), bindingIDLabel, bindingID)
		configs, err := k.configClient.List(typ, k.namespace)
		if err != nil {
			return err
		}
		for _, config := range configs {
			if config.Labels != nil && config.Labels[bindingIDLabel] == bindingID {
				err = k.configClient.Delete(typ, config.Name, k.namespace)
				if err != nil {
					return err
				}
			}
		}

	}
	return nil
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
