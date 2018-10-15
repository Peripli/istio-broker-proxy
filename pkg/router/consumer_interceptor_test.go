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
		Credentials: model.Credentials{
			Endpoints: []model.Endpoint{
				{
					Host: "0.678.services.cf.dev01.aws.istio.sapcloud.io",
					Port: 9001}}}}
	bindResponseTwoEndpoints = model.BindResponse{
		Credentials: model.Credentials{
			Endpoints: []model.Endpoint{
				{
					Host: "0.678.services.cf.dev01.aws.istio.sapcloud.io",
					Port: 9001},
				{
					Host: "1.678.services.cf.dev01.aws.istio.sapcloud.io",
					Port: 9001},
			}}}
)

func TestConsumerPostBind(t *testing.T) {
	g := NewGomegaWithT(t)
	kubernetes := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &kubernetes}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseSingleEndpoint, "678", adapt)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(kubernetes.createdServices[0].Name).To(Equal("svc-0-678"))
	g.Expect(kubernetes.createdServices[0].Spec.Ports[0].Port).To(Equal(int32(5555)))
}

func TestNoEndpointsPresent(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, model.BindResponse{}, "678", adapt)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(configStore.createdServices).To(BeNil())
	g.Expect(configStore.createdIstioConfigs).To(BeNil())
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
			Credentials: model.Credentials{
				Endpoints: endpoints,
			},
			Hostname: "10.10.10.11",
			Port:     5432,
			Uri:      "postgres://user:password@10.10.10.11:5432/test",
		}.ToCredentials(),
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
	g.Expect(configStore.createdServices[0].Name).To(Equal("svc-0-555"))
}

func TestMaximumLengthIsNotExceededWithRealBindId(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseSingleEndpoint, "f1b32107-c8a5-11e8-b8be-02caceffa7f1", adapt)
	g.Expect(err).NotTo(HaveOccurred())

	const maxLabelLength = 63
	for _, object := range configStore.createdIstioConfigs {
		g.Expect(len(object.Name)).To(BeNumerically("<", maxLabelLength), "%s is too long", object.Name)
	}
}

func TestEndpointIndexIsPartOfServiceName(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseTwoEndpoints, "adf123", adapt)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(configStore.createdServices[1].Name).To(Equal("svc-1-adf123"))
}

func TestConsumerInterceptorCreatesIstioObjects(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseSingleEndpoint, "678", adapt)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(len(configStore.createdIstioConfigs)).To(Equal(6))
	g.Expect(configStore.createdIstioConfigs[0].Type).To(Equal("service-entry"))
	text, _ := json.Marshal(configStore.createdIstioConfigs[0])
	g.Expect(text).To(ContainSubstring("0.678.services.cf.dev01.aws.istio.sapcloud.io"))
	g.Expect(configStore.createdIstioConfigs[1].Type).To(Equal("virtual-service"))
	text, _ = json.Marshal(configStore.createdIstioConfigs[1])
	g.Expect(text).To(ContainSubstring("svc-0-678"))
	g.Expect(configStore.createdIstioConfigs[2].Type).To(Equal("virtual-service"))
	g.Expect(configStore.createdIstioConfigs[3].Type).To(Equal("gateway"))
	g.Expect(configStore.createdIstioConfigs[4].Type).To(Equal("destination-rule"))
	g.Expect(configStore.createdIstioConfigs[5].Type).To(Equal("destination-rule"))
}

func TestTwoEndpointsCreateTwelveObject(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseTwoEndpoints, "678", adapt)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(len(configStore.createdIstioConfigs)).To(Equal(12))
	text, err := json.Marshal(configStore.createdIstioConfigs[6])
	g.Expect(text).To(ContainSubstring("1.678.services.cf.dev01.aws.istio.sapcloud.io"))
}

func TestClusterIpIsUsed(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{clusterIp: "9.8.7.6"}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.postBind(model.BindRequest{}, bindResponseTwoEndpoints, "678", adapt)
	g.Expect(err).NotTo(HaveOccurred())

	text, err := json.Marshal(configStore.createdIstioConfigs[1])
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

func TestConsumerPostDelete(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerId: "consumer-id", ConfigStore: &configStore}
	err := consumer.postDelete("678")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(len(configStore.deletedServices)).To(Equal(1))
	g.Expect(len(configStore.deletedIstioConfigs)).To(Equal(6))
	g.Expect(configStore.deletedIstioConfigs[0]).To(Equal("DestinationRule:sidecar-to-egress-svc-0-678"))
	g.Expect(configStore.deletedIstioConfigs[1]).To(Equal("DestinationRule:egressgateway-svc-0-678"))
	g.Expect(configStore.deletedIstioConfigs[2]).To(Equal("Gateway:istio-egressgateway-svc-0-678"))
	g.Expect(configStore.deletedIstioConfigs[3]).To(Equal("VirtualService:egress-gateway-svc-0-678"))
	g.Expect(configStore.deletedIstioConfigs[4]).To(Equal("VirtualService:mesh-to-egress-svc-0-678"))
	g.Expect(configStore.deletedIstioConfigs[5]).To(Equal("ServiceEntry:svc-0-678-service"))
}

type mockConfigStore struct {
	createdServices     []*v1.Service
	createdIstioConfigs []istioModel.Config
	clusterIp           string
	createServiceErr    error
	createObjectErr     error
	deletedServices     []string
	deletedIstioConfigs []string
}

func (m *mockConfigStore) CreateService(service *v1.Service) (*v1.Service, error) {
	if m.createServiceErr != nil {
		return nil, m.createServiceErr
	}
	m.createdServices = append(m.createdServices, service)
	service.Spec.ClusterIP = m.clusterIp
	return service, nil
}

func (m *mockConfigStore) CreateIstioConfig(object istioModel.Config) error {
	if m.createObjectErr != nil {
		return m.createObjectErr
	}
	m.createdIstioConfigs = append(m.createdIstioConfigs, object)
	return nil
}

func (m *mockConfigStore) DeleteService(serviceName string) error {
	m.deletedServices = append(m.deletedServices, serviceName)
	return nil
}

func (m *mockConfigStore) DeleteIstioConfig(configType string, configName string) error {
	m.deletedIstioConfigs = append(m.deletedIstioConfigs, configType+":"+configName)
	return nil
}
