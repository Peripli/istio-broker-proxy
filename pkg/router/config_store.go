package router

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
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
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	namespace, err := getNamespace()
	if err != nil {
		panic(err.Error())
	}
	return newKubeConfigStore(config, namespace)
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

func (k kubeConfigStore) CreateIstioConfig(config model.Config) error {

	schema, exists := model.IstioConfigTypes.GetByType(config.Type)
	if !exists {
		return fmt.Errorf("unrecognized type %q", config.Type)
	}

	if err := schema.Validate(config.Name, config.Namespace, config.Spec); err != nil {
		return multierror.Prefix(err, "validation error:")
	}

	out, err := crd.ConvertConfig(schema, config)
	if err != nil {
		return err
	}

	_, err = k.RESTClient().Post().
		Namespace(out.GetObjectMeta().Namespace).
		Resource(crd.ResourceName(schema.Plural)).
		Body(out).
		Do().
		Get()
	return err
}

func NewExternKubeConfigStore(namespace string) ConfigStore {
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		panic(err.Error())
	}
	return newKubeConfigStore(config, namespace)

}
