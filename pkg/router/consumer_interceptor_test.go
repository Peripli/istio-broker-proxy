package router

import (
	"github.com/gin-gonic/gin/json"
	. "github.com/onsi/gomega"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/profiles"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"os"

	"testing"
)

func TestConsumerPreBind(t *testing.T) {
	g := NewGomegaWithT(t)

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id"}
	request := consumer.preBind(model.BindRequest{})
	g.Expect(request.NetworkData.NetworkProfileId).To(Equal(profiles.NetworkProfile))
	g.Expect(request.NetworkData.Data.ConsumerId).To(Equal("consumer-id"))

}

var (
	bindResponseSingleEndpoint = model.BindResponse{
		NetworkData: model.NetworkDataResponse{
			Data: model.DataResponse{
				Endpoints: []model.Endpoint{
					model.Endpoint{
						Host: "0.678.services.cf.dev01.aws.istio.sapcloud.io",
						Port: 9001}}}}}
	bindResponseTwoEndpoints = model.BindResponse{
		NetworkData: model.NetworkDataResponse{
			Data: model.DataResponse{
				Endpoints: []model.Endpoint{
					model.Endpoint{
						Host: "0.678.services.cf.dev01.aws.istio.sapcloud.io",
						Port: 9001},
					model.Endpoint{
						Host: "1.678.services.cf.dev01.aws.istio.sapcloud.io",
						Port: 9001},
				}}}}
)

func TestConsumerPostBind(t *testing.T) {
	g := NewGomegaWithT(t)
	serviceFactory := mockServiceInterface{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", Kubernetes: &serviceFactory}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseSingleEndpoint, "678")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(serviceFactory.calledWithService[0].Name).To(Equal("service-0-678"))
	g.Expect(serviceFactory.calledWithService[0].Spec.Ports[0].Port).To(Equal(int32(5555)))
}

func TestNoEndpointsPresent(t *testing.T) {
	g := NewGomegaWithT(t)
	serviceFactory := mockServiceInterface{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", Kubernetes: &serviceFactory}
	_, err := consumer.postBind(model.BindRequest{}, model.BindResponse{}, "678")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(serviceFactory.calledWithService).To(BeNil())
	g.Expect(serviceFactory.calledWithObject).To(BeNil())
}

func TestBindIdIsPartOfServiceName(t *testing.T) {
	g := NewGomegaWithT(t)
	serviceFactory := mockServiceInterface{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", Kubernetes: &serviceFactory}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseSingleEndpoint, "555")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(serviceFactory.calledWithService[0].Name).To(Equal("service-0-555"))
}

func TestEndpointIndexIsPartOfServiceName(t *testing.T) {
	g := NewGomegaWithT(t)
	serviceFactory := mockServiceInterface{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", Kubernetes: &serviceFactory}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseTwoEndpoints, "adf123")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(serviceFactory.calledWithService[1].Name).To(Equal("service-1-adf123"))
}

func TestConsumerInterceptorCreatesIstioObjects(t *testing.T) {
	g := NewGomegaWithT(t)
	serviceFactory := mockServiceInterface{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", Kubernetes: &serviceFactory}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseSingleEndpoint, "678")
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(len(serviceFactory.calledWithObject)).To(Equal(6))
	g.Expect(serviceFactory.calledWithObject[0].GetObjectKind().GroupVersionKind().Kind).To(Equal("ServiceEntry"))
	text, _ := json.Marshal(serviceFactory.calledWithObject[0])
	g.Expect(text).To(ContainSubstring("0.678.services.cf.dev01.aws.istio.sapcloud.io"))
	g.Expect(serviceFactory.calledWithObject[1].GetObjectKind().GroupVersionKind().Kind).To(Equal("VirtualService"))
	text, _ = json.Marshal(serviceFactory.calledWithObject[1])
	g.Expect(text).To(ContainSubstring("service-0-678"))
	g.Expect(serviceFactory.calledWithObject[2].GetObjectKind().GroupVersionKind().Kind).To(Equal("VirtualService"))
	g.Expect(serviceFactory.calledWithObject[3].GetObjectKind().GroupVersionKind().Kind).To(Equal("Gateway"))
	g.Expect(serviceFactory.calledWithObject[4].GetObjectKind().GroupVersionKind().Kind).To(Equal("DestinationRule"))
	g.Expect(serviceFactory.calledWithObject[5].GetObjectKind().GroupVersionKind().Kind).To(Equal("DestinationRule"))
}

func TestTwoEndpointsCreateTwelveObject(t *testing.T) {
	g := NewGomegaWithT(t)
	serviceFactory := mockServiceInterface{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", Kubernetes: &serviceFactory}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseTwoEndpoints, "678")
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(len(serviceFactory.calledWithObject)).To(Equal(12))
	text, err := json.Marshal(serviceFactory.calledWithObject[6])
	g.Expect(text).To(ContainSubstring("1.678.services.cf.dev01.aws.istio.sapcloud.io"))
}

type mockServiceInterface struct {
	calledWithService []*v1.Service
	calledWithObject  []runtime.Object
}

func (m *mockServiceInterface) CreateService(service *v1.Service) (*v1.Service, error) {
	m.calledWithService = append(m.calledWithService, service)
	//service.Spec.ClusterIP = "9.9.9.9"
	return service, nil
}

func (m *mockServiceInterface) CreateObject(object runtime.Object) error {
	m.calledWithObject = append(m.calledWithObject, object)
	return nil
}

func KubeConfigServiceFactory() clientv1.ServiceInterface {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset.CoreV1().Services("cki75")

}
