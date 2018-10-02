package router

import (
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/config"
	"io/ioutil"
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

	//config.GroupVersion = &schema.GroupVersion{
	//	Group:   "config.istio.io",
	//	Version: "v1alpha2",
	//}
	//config.APIPath = "/apis"
	//config.ContentType = runtime.ContentTypeJSON
	//
	//types := runtime.NewScheme()
	//schemeBuilder := runtime.NewSchemeBuilder(
	//	func(scheme *runtime.Scheme) error {
	//		metav1.AddToGroupVersion(scheme, *config.GroupVersion)
	//		return nil
	//	})
	//err := schemeBuilder.AddToScheme(types)
	//config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(types)}
	//
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

	_, err := config.ToRuntimeObject(cfg)

	//_, err = k.RESTClient().Post().
	//	Namespace(out.GetObjectMeta().Namespace).
	//	Resource(crd.ResourceName(schema.Plural)).
	//	Body(out).
	//	Do().
	//	Get()
	return err
}

func NewExternKubeConfigStore(namespace string) ConfigStore {
	cfg, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		panic(err.Error())
	}
	return newKubeConfigStore(cfg, namespace)

}
