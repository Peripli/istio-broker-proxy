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

func adapt(credentials model.Credentials, endpointMappings []model.EndpointMapping) (*model.BindResponse, error) {
	return &model.BindResponse{Credentials: credentials}, nil
}

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
	_, err := consumer.postBind(model.BindRequest{}, bindResponseSingleEndpoint, "678", adapt)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(kubernetes.calledWithService[0].Name).To(Equal("svc-0-678"))
	g.Expect(kubernetes.calledWithService[0].Spec.Ports[0].Port).To(Equal(int32(5555)))
}

func TestNoEndpointsPresent(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, model.BindResponse{}, "678", adapt)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(configStore.calledWithService).To(BeNil())
	g.Expect(configStore.calledWithObject).To(BeNil())
}

func TestEndpointsMappingWorks(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore, SystemDomain: "cluster.local", Namespace: "catalog"}
	endpoints := []model.Endpoint{
		{
			Host: "10.10.10.11",
			Port: 5432,
		},
	}
	binding, err := consumer.postBind(model.BindRequest{},
		model.BindResponse{Credentials: model.PostgresCredentials{
			Hostname: "10.10.10.11",
			Port:     5432,
			Uri:      "postgres://user:password@10.10.10.11:5432/test",
		}.ToCredentials(),
			Endpoints:   endpoints,
			NetworkData: model.NetworkDataResponse{NetworkProfileId: "testprofile", Data: model.DataResponse{Endpoints: endpoints}}},
		"678", model.Adapt)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(binding.Endpoints).NotTo(BeNil())
	g.Expect(len(binding.Endpoints)).To(Equal(1))
	g.Expect(binding.Endpoints[0].Host).To(Equal("svc-0-678.catalog.svc.cluster.local"))
	g.Expect(binding.Endpoints[0].Port).To(Equal(service_port))
	g.Expect(binding.NetworkData.NetworkProfileId).To(Equal("testprofile"))
	g.Expect(len(binding.NetworkData.Data.Endpoints)).To(Equal(1))
	g.Expect(binding.NetworkData.Data.Endpoints[0]).To(Equal(endpoints[0]))

	postgresCredentials, err := model.PostgresCredentialsFromCredentials(binding.Credentials)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(postgresCredentials.Hostname).To(Equal("svc-0-678.catalog.svc.cluster.local"))
	g.Expect(postgresCredentials.Port).To(Equal(service_port))
	g.Expect(postgresCredentials.Uri).To(Equal("postgres://user:password@svc-0-678.catalog.svc.cluster.local:5555/test"))

}

func TestBindIdIsPartOfServiceName(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseSingleEndpoint, "555", adapt)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(configStore.calledWithService[0].Name).To(Equal("svc-0-555"))
}

func TestMaximumLengthIsNotExceededWithRealBindId(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseSingleEndpoint, "f1b32107-c8a5-11e8-b8be-02caceffa7f1", adapt)
	g.Expect(err).NotTo(HaveOccurred())

	const maxLabelLength = 63
	for _, object := range configStore.calledWithObject {
		g.Expect(len(object.Name)).To(BeNumerically("<", maxLabelLength), "%s is too long", object.Name)
	}
}

func TestEndpointIndexIsPartOfServiceName(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseTwoEndpoints, "adf123", adapt)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(configStore.calledWithService[1].Name).To(Equal("svc-1-adf123"))
}

func TestConsumerInterceptorCreatesIstioObjects(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseSingleEndpoint, "678", adapt)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(len(configStore.calledWithObject)).To(Equal(6))
	g.Expect(configStore.calledWithObject[0].Type).To(Equal("service-entry"))
	text, _ := json.Marshal(configStore.calledWithObject[0])
	g.Expect(text).To(ContainSubstring("0.678.services.cf.dev01.aws.istio.sapcloud.io"))
	g.Expect(configStore.calledWithObject[1].Type).To(Equal("virtual-service"))
	text, _ = json.Marshal(configStore.calledWithObject[1])
	g.Expect(text).To(ContainSubstring("svc-0-678"))
	g.Expect(configStore.calledWithObject[2].Type).To(Equal("virtual-service"))
	g.Expect(configStore.calledWithObject[3].Type).To(Equal("gateway"))
	g.Expect(configStore.calledWithObject[4].Type).To(Equal("destination-rule"))
	g.Expect(configStore.calledWithObject[5].Type).To(Equal("destination-rule"))
}

func TestTwoEndpointsCreateTwelveObject(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseTwoEndpoints, "678", adapt)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(len(configStore.calledWithObject)).To(Equal(12))
	text, err := json.Marshal(configStore.calledWithObject[6])
	g.Expect(text).To(ContainSubstring("1.678.services.cf.dev01.aws.istio.sapcloud.io"))
}

func TestClusterIpIsUsed(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{clusterIp: "9.8.7.6"}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseTwoEndpoints, "678", adapt)
	g.Expect(err).NotTo(HaveOccurred())

	text, err := json.Marshal(configStore.calledWithObject[1])
	g.Expect(text).To(ContainSubstring("9.8.7.6"))
}

func TestCreateServiceErrorIsHandled(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{createServiceErr: fmt.Errorf("Test service error")}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseTwoEndpoints, "678", adapt)
	g.Expect(err.Error()).To(Equal("Test service error"))
}

func TestCreateObjectErrorIsHandled(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{createObjectErr: fmt.Errorf("Test object error")}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseTwoEndpoints, "678", adapt)
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
