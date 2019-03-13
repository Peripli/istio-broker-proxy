package router

import (
	"errors"
	"fmt"
	"github.com/Peripli/istio-broker-proxy/pkg/model"
	"github.com/gin-gonic/gin/json"
	. "github.com/onsi/gomega"
	"testing"
)

func adapt(credentials model.Credentials, endpointMappings []model.EndpointMapping) (*model.BindResponse, error) {
	return &model.BindResponse{Credentials: credentials}, nil
}

func adaptError(credentials model.Credentials, endpointMappings []model.EndpointMapping) (*model.BindResponse, error) {
	return &model.BindResponse{Credentials: credentials}, errors.New("error during adapt")
}

func TestConsumerPreBind(t *testing.T) {
	g := NewGomegaWithT(t)

	consumer := ConsumerInterceptor{ConsumerID: "consumer-id", NetworkProfile: "network-profile"}
	request, err := consumer.PreBind(model.BindRequest{})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(request.NetworkData.NetworkProfileID).To(Equal("network-profile"))
	g.Expect(request.NetworkData.Data.ConsumerID).To(Equal("consumer-id"))

}

func TestConsumerPreBindWithoutNetworkProfile(t *testing.T) {
	g := NewGomegaWithT(t)

	consumer := ConsumerInterceptor{ConsumerID: "consumer-id"}
	_, err := consumer.PreBind(model.BindRequest{})
	g.Expect(err).To(HaveOccurred())
}

var (
	bindResponseSingleEndpoint = model.BindResponse{
		Credentials: model.Credentials{},
		Endpoints: []model.Endpoint{
			{
				Host: "10.11.12.13",
				Port: 5432},
		},
		NetworkData: model.NetworkDataResponse{Data: model.DataResponse{
			Endpoints: []model.Endpoint{
				{
					Host: "0.678.istio.my.arbitrary.domain.io",
					Port: 9001}}}}}
	bindResponseTwoEndpoints = model.BindResponse{
		Credentials: model.Credentials{},
		Endpoints: []model.Endpoint{
			{
				Host: "10.11.12.13",
				Port: 5432},
			{
				Host: "10.11.12.13",
				Port: 5432},
		},
		NetworkData: model.NetworkDataResponse{Data: model.DataResponse{
			Endpoints: []model.Endpoint{
				{
					Host: "0.678.istio.my.arbitrary.domain.io",
					Port: 9001},
				{
					Host: "1.678.istio.my.arbitrary.domain.io",
					Port: 9001},
			},
		}}}
)

func TestConsumerPostBind(t *testing.T) {
	g := NewGomegaWithT(t)
	kubernetes := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerID: "consumer-id", ConfigStore: &kubernetes}
	_, err := consumer.PostBind(model.BindRequest{}, bindResponseSingleEndpoint, "678", adapt)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(kubernetes.CreatedServices).To(HaveLen(1))
	g.Expect(kubernetes.CreatedServices[0].Name).To(Equal("svc-0-678"))
	g.Expect(kubernetes.CreatedServices[0].Spec.Ports[0].Port).To(Equal(int32(5555)))
}

func TestConsumerPostBindReturnsError(t *testing.T) {
	g := NewGomegaWithT(t)
	kubernetes := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerID: "consumer-id", ConfigStore: &kubernetes}
	_, err := consumer.PostBind(model.BindRequest{}, bindResponseSingleEndpoint, "678", adaptError)
	g.Expect(err).To(HaveOccurred())
}

func TestNoEndpointsPresent(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerID: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.PostBind(model.BindRequest{}, model.BindResponse{}, "678", adapt)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(configStore.CreatedServices).To(BeNil())
	g.Expect(configStore.CreatedIstioConfigs).To(BeNil())
}

func TestEndpointsMappingWorks(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}
	configStore.ClusterIP = "1.2.3.5"
	consumer := ConsumerInterceptor{ConsumerID: "consumer-id", ConfigStore: &configStore, NetworkProfile: "testprofile"}
	endpoints := []model.Endpoint{
		{
			Host: "10.10.10.11",
			Port: 5432,
		},
	}
	binding, err := consumer.PostBind(model.BindRequest{},
		model.BindResponse{Credentials: model.PostgresCredentials{
			Credentials: model.Credentials{},
			Hostname:    "10.10.10.11",
			Port:        5432,
			URI:         "postgres://user:password@10.10.10.11:5432/test",
		}.ToCredentials(),
			Endpoints:   endpoints,
			NetworkData: model.NetworkDataResponse{NetworkProfileID: "testprofile", Data: model.DataResponse{Endpoints: endpoints}}},
		"678", model.Adapt)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(binding.Endpoints).NotTo(BeNil())
	g.Expect(len(binding.Endpoints)).To(Equal(1))
	g.Expect(binding.Endpoints[0].Host).To(Equal("1.2.3.5"))
	g.Expect(binding.Endpoints[0].Port).To(Equal(servicePort))
	g.Expect(binding.NetworkData.NetworkProfileID).To(Equal("testprofile"))
	g.Expect(len(binding.NetworkData.Data.Endpoints)).To(Equal(1))
	g.Expect(binding.NetworkData.Data.Endpoints[0]).To(Equal(endpoints[0]))

	postgresCredentials, err := model.PostgresCredentialsFromCredentials(binding.Credentials)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(postgresCredentials.Hostname).To(Equal("1.2.3.5"))
	g.Expect(postgresCredentials.Port).To(Equal(servicePort))
	g.Expect(postgresCredentials.URI).To(Equal("postgres://user:password@1.2.3.5:5555/test"))

}

func TestBindIdIsPartOfServiceName(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerID: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.PostBind(model.BindRequest{}, bindResponseSingleEndpoint, "555", adapt)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(configStore.CreatedServices).To(HaveLen(1))
	g.Expect(configStore.CreatedServices[0].Name).To(Equal("svc-0-555"))
}

func TestMaximumLengthIsNotExceededWithRealBindId(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerID: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.PostBind(model.BindRequest{}, bindResponseSingleEndpoint, "f1b32107-c8a5-11e8-b8be-02caceffa7f1", adapt)
	g.Expect(err).NotTo(HaveOccurred())

	const maxLabelLength = 63
	for _, object := range configStore.CreatedIstioConfigs {
		g.Expect(len(object.Name)).To(BeNumerically("<", maxLabelLength), "%s is too long", object.Name)
	}
}

func TestEndpointIndexIsPartOfServiceName(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerID: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.PostBind(model.BindRequest{}, bindResponseTwoEndpoints, "adf123", adapt)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(configStore.CreatedServices).To(HaveLen(2))
	g.Expect(configStore.CreatedServices[1].Name).To(Equal("svc-1-adf123"))
}

func TestConsumerInterceptorCreatesIstioObjects(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerID: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.PostBind(model.BindRequest{}, bindResponseSingleEndpoint, "678", adapt)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(len(configStore.CreatedIstioConfigs)).To(Equal(6))
	g.Expect(configStore.CreatedIstioConfigs[0].Type).To(Equal("service-entry"))
	text, _ := json.Marshal(configStore.CreatedIstioConfigs[0])
	g.Expect(text).To(ContainSubstring("0.678.istio.my.arbitrary.domain.io"))
	g.Expect(configStore.CreatedIstioConfigs[1].Type).To(Equal("virtual-service"))
	text, _ = json.Marshal(configStore.CreatedIstioConfigs[1])
	g.Expect(text).To(ContainSubstring("svc-0-678"))
	g.Expect(configStore.CreatedIstioConfigs[2].Type).To(Equal("virtual-service"))
	g.Expect(configStore.CreatedIstioConfigs[3].Type).To(Equal("gateway"))
	g.Expect(configStore.CreatedIstioConfigs[4].Type).To(Equal("destination-rule"))
	g.Expect(configStore.CreatedIstioConfigs[5].Type).To(Equal("destination-rule"))
}

func TestTwoEndpointsCreateTwelveObject(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerID: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.PostBind(model.BindRequest{}, bindResponseTwoEndpoints, "678", adapt)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(len(configStore.CreatedIstioConfigs)).To(Equal(12))
	text, err := json.Marshal(configStore.CreatedIstioConfigs[6])
	g.Expect(text).To(ContainSubstring("1.678.istio.my.arbitrary.domain.io"))
}

func TestTwoEndpointsHasTheCorrcetCount(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}
	bindResponse := model.BindResponse{
		Credentials: model.Credentials{},
		Endpoints:   []model.Endpoint{},
		NetworkData: model.NetworkDataResponse{Data: model.DataResponse{
			Endpoints: []model.Endpoint{
				{
					Host: "0.678.istio.my.arbitrary.domain.io",
					Port: 9001}}}}}

	consumer := ConsumerInterceptor{ConsumerID: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.PostBind(model.BindRequest{}, bindResponse, "678", adapt)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("Number of endpoints"))
}

func TestClusterIpIsUsed(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{ClusterIP: "9.8.7.6"}

	consumer := ConsumerInterceptor{ConsumerID: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.PostBind(model.BindRequest{}, bindResponseTwoEndpoints, "678", adapt)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(configStore.CreatedIstioConfigs).To(HaveLen(12))
	text, err := json.Marshal(configStore.CreatedIstioConfigs[1])
	g.Expect(text).To(ContainSubstring("9.8.7.6"))
}

func TestCreateServiceErrorIsHandled(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{CreateServiceErr: fmt.Errorf("Test service error")}

	consumer := ConsumerInterceptor{ConsumerID: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.PostBind(model.BindRequest{}, bindResponseTwoEndpoints, "678", adapt)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(Equal("Test service error"))
}

func TestCreateObjectErrorIsHandled(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{CreateObjectErr: fmt.Errorf("Test object error")}

	consumer := ConsumerInterceptor{ConsumerID: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.PostBind(model.BindRequest{}, bindResponseTwoEndpoints, "678", adapt)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(Equal("Test object error"))
}

func TestConsumerPostDelete(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerID: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.PostBind(model.BindRequest{}, bindResponseTwoEndpoints, "678", adapt)
	g.Expect(err).NotTo(HaveOccurred())

	consumer.PostDelete("678")
	g.Expect(len(configStore.DeletedServices)).To(Equal(2))
	g.Expect(len(configStore.DeletedIstioConfigs)).To(Equal(12))
	g.Expect(configStore.DeletedIstioConfigs[0]).To(Equal("destination-rule:sidecar-to-egress-svc-0-678"))
	g.Expect(configStore.DeletedIstioConfigs[1]).To(Equal("destination-rule:egressgateway-svc-0-678"))
	g.Expect(configStore.DeletedIstioConfigs[2]).To(Equal("gateway:istio-egressgateway-svc-0-678"))
	g.Expect(configStore.DeletedIstioConfigs[3]).To(Equal("virtual-service:egress-gateway-svc-0-678"))
	g.Expect(configStore.DeletedIstioConfigs[4]).To(Equal("virtual-service:mesh-to-egress-svc-0-678"))
	g.Expect(configStore.DeletedIstioConfigs[5]).To(Equal("service-entry:svc-0-678-service"))
	g.Expect(configStore.DeletedIstioConfigs[6]).To(Equal("destination-rule:sidecar-to-egress-svc-1-678"))
	g.Expect(configStore.DeletedIstioConfigs[7]).To(Equal("destination-rule:egressgateway-svc-1-678"))
	g.Expect(configStore.DeletedIstioConfigs[8]).To(Equal("gateway:istio-egressgateway-svc-1-678"))
	g.Expect(configStore.DeletedIstioConfigs[9]).To(Equal("virtual-service:egress-gateway-svc-1-678"))
	g.Expect(configStore.DeletedIstioConfigs[10]).To(Equal("virtual-service:mesh-to-egress-svc-1-678"))
	g.Expect(configStore.DeletedIstioConfigs[11]).To(Equal("service-entry:svc-1-678-service"))
}

func TestConsumerPostDeleteNoResourceLeaks(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerID: "consumer-id", ConfigStore: &configStore}
	_, err := consumer.PostBind(model.BindRequest{}, bindResponseTwoEndpoints, "678", adapt)

	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(configStore.CreatedServices).To(HaveLen(2))
	g.Expect(configStore.CreatedIstioConfigs).To(HaveLen(12))

	configStore.CreatedServices = configStore.CreatedServices[1:]
	configStore.CreatedIstioConfigs = append(configStore.CreatedIstioConfigs[0:3], configStore.CreatedIstioConfigs[5:]...)

	consumer.PostDelete("678")

	g.Expect(configStore.CreatedServices).To(HaveLen(0))
	g.Expect(configStore.CreatedIstioConfigs).To(HaveLen(0))
}

func TestConsumerFailingPostBindGetsCleanedUp(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{CreateObjectErrCount: 3, CreateObjectErr: errors.New("No more objects")}

	consumer := ConsumerInterceptor{ConsumerID: "consumer-id", ConfigStore: &configStore}

	_, err := consumer.PostBind(model.BindRequest{}, bindResponseTwoEndpoints, "678", adapt)
	g.Expect(err).To(HaveOccurred())

	g.Expect(configStore.CreatedServices).To(HaveLen(0))
	g.Expect(configStore.CreatedIstioConfigs).To(HaveLen(0))
}

func TestConsumerPostCatalog(t *testing.T) {
	g := NewGomegaWithT(t)
	interceptor := ConsumerInterceptor{ServiceNamePrefix: "istio-"}
	catalog := model.Catalog{[]model.Service{{Name: "istio-name"}}}
	interceptor.PostCatalog(&catalog)
	g.Expect(catalog.Services[0].Name).To(Equal("name"))

}

func TestConsumerPostCatalogWithoutPrefix(t *testing.T) {
	g := NewGomegaWithT(t)
	interceptor := ConsumerInterceptor{ServiceNamePrefix: "istio-"}
	catalog := model.Catalog{[]model.Service{{Name: "test-xxx-name"}}}
	interceptor.PostCatalog(&catalog)
	g.Expect(catalog.Services[0].Name).To(Equal("test-xxx-name"))

}

func TestBindAdaptEndpointsOnlyIfNetworkProfilesMatch(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := mockConfigStore{}

	consumer := ConsumerInterceptor{ConsumerID: "consumer-id", ConfigStore: &configStore, NetworkProfile: "urn:my.test:public"}
	binding, err := consumer.PostBind(model.BindRequest{}, bindResponseSingleEndpoint, "555", adaptError)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(*binding).To(Equal(bindResponseSingleEndpoint))
}
