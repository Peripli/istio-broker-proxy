package router

import (
	"fmt"
	"github.com/gin-gonic/gin/json"
	. "github.com/onsi/gomega"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/profiles"
	istioModel "istio.io/istio/pilot/pkg/model"
	"k8s.io/api/core/v1"
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
					{
						Host: "0.678.services.cf.dev01.aws.istio.sapcloud.io",
						Port: 9001}}}}}
	bindResponseTwoEndpoints = model.BindResponse{
		NetworkData: model.NetworkDataResponse{
			Data: model.DataResponse{
				Endpoints: []model.Endpoint{
					{
						Host: "0.678.services.cf.dev01.aws.istio.sapcloud.io",
						Port: 9001},
					{
						Host: "1.678.services.cf.dev01.aws.istio.sapcloud.io",
						Port: 9001},
				}}}}
)

func TestConsumerPostBind(t *testing.T) {
	g := NewGomegaWithT(t)
	kubernetes := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &kubernetes}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseSingleEndpoint, "678")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(kubernetes.calledWithService[0].Name).To(Equal("service-0-678"))
	g.Expect(kubernetes.calledWithService[0].Spec.Ports[0].Port).To(Equal(int32(5555)))
}

func TestNoEndpointsPresent(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, model.BindResponse{}, "678")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(configStore.calledWithService).To(BeNil())
	g.Expect(configStore.calledWithObject).To(BeNil())
}

func TestBindIdIsPartOfServiceName(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseSingleEndpoint, "555")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(configStore.calledWithService[0].Name).To(Equal("service-0-555"))
}

func TestEndpointIndexIsPartOfServiceName(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseTwoEndpoints, "adf123")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(configStore.calledWithService[1].Name).To(Equal("service-1-adf123"))
}

func TestConsumerInterceptorCreatesIstioObjects(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseSingleEndpoint, "678")
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(len(configStore.calledWithObject)).To(Equal(6))
	g.Expect(configStore.calledWithObject[0].Type).To(Equal("service-entry"))
	text, _ := json.Marshal(configStore.calledWithObject[0])
	g.Expect(text).To(ContainSubstring("0.678.services.cf.dev01.aws.istio.sapcloud.io"))
	g.Expect(configStore.calledWithObject[1].Type).To(Equal("virtual-service"))
	text, _ = json.Marshal(configStore.calledWithObject[1])
	g.Expect(text).To(ContainSubstring("service-0-678"))
	g.Expect(configStore.calledWithObject[2].Type).To(Equal("virtual-service"))
	g.Expect(configStore.calledWithObject[3].Type).To(Equal("gateway"))
	g.Expect(configStore.calledWithObject[4].Type).To(Equal("destination-rule"))
	g.Expect(configStore.calledWithObject[5].Type).To(Equal("destination-rule"))
}

func TestTwoEndpointsCreateTwelveObject(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseTwoEndpoints, "678")
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(len(configStore.calledWithObject)).To(Equal(12))
	text, err := json.Marshal(configStore.calledWithObject[6])
	g.Expect(text).To(ContainSubstring("1.678.services.cf.dev01.aws.istio.sapcloud.io"))
}

func TestClusterIpIsUsed(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{clusterIp: "9.8.7.6"}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseTwoEndpoints, "678")
	g.Expect(err).NotTo(HaveOccurred())

	text, err := json.Marshal(configStore.calledWithObject[1])
	g.Expect(text).To(ContainSubstring("9.8.7.6"))
}

func TestCreateServiceErrorIsHandled(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{createServiceErr: fmt.Errorf("Test service error")}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseTwoEndpoints, "678")
	g.Expect(err.Error()).To(Equal("Test service error"))
}

func TestCreateObjectErrorIsHandled(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{createObjectErr: fmt.Errorf("Test object error")}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseTwoEndpoints, "678")
	g.Expect(err.Error()).To(Equal("Test object error"))
}

type mockConfigStore struct {
	calledWithService []*v1.Service
	calledWithObject  []istioModel.Config
	clusterIp         string
	createServiceErr  error
	createObjectErr   error
}

func (m *mockConfigStore) CreateService(service *v1.Service) (*v1.Service, error) {
	if m.createServiceErr != nil {
		return nil, m.createServiceErr
	}
	m.calledWithService = append(m.calledWithService, service)
	service.Spec.ClusterIP = m.clusterIp
	return service, nil
}

func (m *mockConfigStore) CreateIstioConfig(object istioModel.Config) error {
	if m.createObjectErr != nil {
		return m.createObjectErr
	}
	m.calledWithObject = append(m.calledWithObject, object)
	return nil
}
