package router

import (
	"encoding/json"
	"github.com/Peripli/istio-broker-proxy/pkg/model"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestDefaultConfigurationIsWritten(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := &fileConfigStore{istioDirectory: os.TempDir()}
	interceptor := ProducerInterceptor{
		ProviderID:   "your-provider",
		SystemDomain: "services.domain",
		ConfigStore:  configStore,
	}
	interceptor.WriteIstioConfigFiles(147)
	file, err := os.Open(path.Join(configStore.istioDirectory, "istio-broker.yml"))
	g.Expect(err).NotTo(HaveOccurred())
	content, err := ioutil.ReadAll(file)
	contentAsString := string(content)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(contentAsString).To(ContainSubstring("147"))
	g.Expect(contentAsString).To(ContainSubstring("istio-broker.services.domain"))
	g.Expect(contentAsString).To(MatchRegexp("number: 9000"))

}

func TestWriteIstioConfigFilesReturnsError(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := &fileConfigStore{istioDirectory: "/not-existing"}
	interceptor := ProducerInterceptor{
		ProviderID:   "your-provider",
		SystemDomain: "services.domain",
		ConfigStore:  configStore}
	err := interceptor.WriteIstioConfigFiles(147)
	g.Expect(err).To(HaveOccurred())
}

func TestYmlFileIsCorrectlyWritten(t *testing.T) {
	g := NewGomegaWithT(t)
	///var/vcap/packages/istio-broker/bin/istio-broker --port 8000 --forwardUrl https://10.11.252.10:9293/cf
	// --systemdomain istio.my.arbitrary.domain.io --ProviderID cf-service.istio.my.arbitrary.domain.io
	// --LoadBalancerPort 9000 --istioDirectory /var/vcap/store/istio-config --ipAddress 10.0.81.0
	configStore := &fileConfigStore{istioDirectory: os.TempDir()}
	interceptor := ProducerInterceptor{
		ProviderID:       "prefix.istio.my.arbitrary.domain.io",
		SystemDomain:     "istio.my.arbitrary.domain.io",
		LoadBalancerPort: 9000,
		IPAddress:        "10.0.81.0",
		ConfigStore:      configStore}
	interceptor.WriteIstioConfigFiles(8000)

	file, err := os.Open(path.Join(configStore.istioDirectory, "istio-broker.yml"))
	g.Expect(err).NotTo(HaveOccurred())
	content, err := ioutil.ReadAll(file)
	contentAsString := string(content)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(contentAsString).To(ContainSubstring("8000"))
	g.Expect(contentAsString).To(ContainSubstring("istio-broker.istio.my.arbitrary.domain.io"))
	g.Expect(contentAsString).To(ContainSubstring("/etc/istio/certs/cf-service.crt"))
	g.Expect(contentAsString).To(MatchRegexp("number: 9000"))

}

func TestEndpointsAreTransferedFromCredentials(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := &fileConfigStore{istioDirectory: os.TempDir()}
	interceptor := ProducerInterceptor{
		ProviderID:       "cf-service.istio.my.arbitrary.domain.io",
		SystemDomain:     "istio.my.arbitrary.domain.io",
		LoadBalancerPort: 9000,
		IPAddress:        "10.0.81.0",
		ConfigStore:      configStore,
		NetworkProfile:   "urn:local.test:public",
	}
	endpoints := []model.Endpoint{{"test.local", 5757}}
	bindResponse, err := interceptor.PostBind(model.BindRequest{}, model.BindResponse{
		Credentials: model.Credentials{
			Endpoints: endpoints,
		},
	}, "123", adapt)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(bindResponse.Endpoints).To(Equal(endpoints))
	g.Expect(len(bindResponse.Credentials.Endpoints)).To(Equal(0))
}

func TestConfigFilesAreWrittenAndDeleted(t *testing.T) {
	g := NewGomegaWithT(t)
	configStore := &fileConfigStore{istioDirectory: os.TempDir()}
	interceptor := ProducerInterceptor{
		ProviderID:       "cf-service.istio.my.arbitrary.domain.io",
		SystemDomain:     "istio.my.arbitrary.domain.io",
		LoadBalancerPort: 9000,
		IPAddress:        "10.0.81.0",
		ConfigStore:      configStore,
		NetworkProfile:   "urn:local.test:public",
	}
	endpoints := []model.Endpoint{{"test.local", 5757}}
	_, err := interceptor.PostBind(model.BindRequest{}, model.BindResponse{
		Credentials: model.Credentials{
			Endpoints: endpoints,
		},
	}, "123", adapt)
	g.Expect(err).NotTo(HaveOccurred())
	fileName := path.Join(configStore.istioDirectory, "123.yml")
	file, err := os.Open(fileName)
	g.Expect(err).NotTo(HaveOccurred())
	content, err := ioutil.ReadAll(file)
	g.Expect(err).NotTo(HaveOccurred())
	contentAsString := string(content)
	g.Expect(contentAsString).To(ContainSubstring("9000"))

	interceptor.PostUnbind("123")
	_, err = os.Stat(fileName)
	g.Expect(err).To(HaveOccurred())

}

func TestConfigFilesBindFailsButFileIsCleanedUp(t *testing.T) {
	g := NewGomegaWithT(t)
	tempDir := os.TempDir()
	configStore := &fileConfigStore{istioDirectory: tempDir}
	interceptor := ProducerInterceptor{
		ProviderID:       "cf-service.istio.my.arbitrary.domain.io",
		SystemDomain:     "istio.my.arbitrary.domain.io",
		LoadBalancerPort: 9000,
		IPAddress:        "10.0.81.0",
		ConfigStore:      configStore,
		NetworkProfile:   "urn:local.test:public",
	}
	endpoints := []model.Endpoint{{"test.local", 5757}}

	//Create file upfront to provoke error
	fileName := path.Join(tempDir, "cant_be_accessed.yml")
	err := os.Mkdir(fileName, os.ModeDir)
	defer os.Remove(fileName)
	g.Expect(err).NotTo(HaveOccurred())

	_, err = interceptor.PostBind(model.BindRequest{}, model.BindResponse{
		Credentials: model.Credentials{
			Endpoints: endpoints,
		},
	}, "cant_be_accessed", adapt)
	g.Expect(err).To(HaveOccurred())

	//file should be deleted
	_, err = os.Stat(fileName)
	g.Expect(err).To(HaveOccurred())

}

func TestProducerPostCatalog(t *testing.T) {
	g := NewGomegaWithT(t)
	interceptor := ProducerInterceptor{ServiceNamePrefix: "istio-"}
	catalog := model.Catalog{[]model.Service{{Name: "name"}}}
	interceptor.PostCatalog(&catalog)
	g.Expect(catalog.Services[0].Name).To(Equal("istio-name"))
}

func TestEnrichPlanMetaData(t *testing.T) {
	g := NewGomegaWithT(t)
	interceptor := ProducerInterceptor{PlanMetaData: `{"supportedPlatforms": ["kubernetes"]}`}
	catalog := model.Catalog{[]model.Service{{Plans: []model.Plan{model.Plan{}, model.Plan{}}}}}
	err := interceptor.PostCatalog(&catalog)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(catalog.Services[0].Plans[0].MetaData["supportedPlatforms"])).To(MatchJSON(`["kubernetes"]`))
	g.Expect(string(catalog.Services[0].Plans[1].MetaData["supportedPlatforms"])).To(MatchJSON(`["kubernetes"]`))
}

func TestServiceWithoutPlanDoesNotLeadToCrash(t *testing.T) {
	g := NewGomegaWithT(t)
	interceptor := ProducerInterceptor{PlanMetaData: `{}`}
	catalog := model.Catalog{[]model.Service{{}}}
	err := interceptor.PostCatalog(&catalog)
	g.Expect(err).NotTo(HaveOccurred())
}

func TestEmptyServiceMetaDataDoesntCrash(t *testing.T) {
	g := NewGomegaWithT(t)
	interceptor := ProducerInterceptor{PlanMetaData: ""}
	catalog := model.Catalog{[]model.Service{{Name: "name"}}}
	err := interceptor.PostCatalog(&catalog)
	g.Expect(err).NotTo(HaveOccurred())
}

func TestEnrichNonEmptyMetaData(t *testing.T) {
	g := NewGomegaWithT(t)
	interceptor := ProducerInterceptor{PlanMetaData: `{"supportedPlatforms": ["kubernetes"]}`}
	catalog := model.Catalog{[]model.Service{{Plans: []model.Plan{{MetaData: map[string]json.RawMessage{"testKey": json.RawMessage(`"testvalue"`)}}}}}}
	err := interceptor.PostCatalog(&catalog)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(catalog.Services[0].Plans[0].MetaData["supportedPlatforms"])).To(MatchJSON(`["kubernetes"]`))
	g.Expect(string(catalog.Services[0].Plans[0].MetaData["testKey"])).To(MatchJSON(`"testvalue"`))
}

func TestEnrichInvalidMetaData(t *testing.T) {
	g := NewGomegaWithT(t)
	interceptor := ProducerInterceptor{PlanMetaData: `{"supportedPlatforms": "invalidJson"]}`}
	catalog := model.Catalog{[]model.Service{{Plans: []model.Plan{model.Plan{}}}}}
	err := interceptor.PostCatalog(&catalog)
	g.Expect(err).To(HaveOccurred())
}
