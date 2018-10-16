package router

import (
	. "github.com/onsi/gomega"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestDefaultConfigurationIsWritten(t *testing.T) {
	g := NewGomegaWithT(t)
	interceptor := ProducerInterceptor{
		ProviderId:     "your-provider",
		SystemDomain:   "services.domain",
		IstioDirectory: os.TempDir()}
	interceptor.WriteIstioConfigFiles(147)
	file, err := os.Open(path.Join(interceptor.IstioDirectory, "istio-broker.yml"))
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
	interceptor := ProducerInterceptor{
		ProviderId:     "your-provider",
		SystemDomain:   "services.domain",
		IstioDirectory: "/not-existing"}
	err := interceptor.WriteIstioConfigFiles(147)
	g.Expect(err).To(HaveOccurred())
}

func TestYmlFileIsCorrectlyWritten(t *testing.T) {
	g := NewGomegaWithT(t)
	///var/vcap/packages/istio-broker/bin/istio-broker --port 8000 --forwardUrl https://10.11.252.10:9293/cf
	// --systemdomain services.cf.dev01.aws.istio.sapcloud.io --ProviderId pinger.services.cf.dev01.aws.istio.sapcloud.io
	// --LoadBalancerPort 9000 --istioDirectory /var/vcap/store/istio-config --ipAddress 10.0.81.0
	interceptor := ProducerInterceptor{
		ProviderId:       "pinger.services.cf.dev01.aws.istio.sapcloud.io",
		SystemDomain:     "services.cf.dev01.aws.istio.sapcloud.io",
		LoadBalancerPort: 9000,
		IpAddress:        "10.0.81.0",
		IstioDirectory:   os.TempDir(),
	}
	interceptor.WriteIstioConfigFiles(8000)

	file, err := os.Open(path.Join(interceptor.IstioDirectory, "istio-broker.yml"))
	g.Expect(err).NotTo(HaveOccurred())
	content, err := ioutil.ReadAll(file)
	contentAsString := string(content)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(contentAsString).To(ContainSubstring("8000"))
	g.Expect(contentAsString).To(ContainSubstring("istio-broker.services.cf.dev01.aws.istio.sapcloud.io"))
	g.Expect(contentAsString).To(MatchRegexp("number: 9000"))

}

func TestEndpointsAreTransferedFromCredentials(t *testing.T) {
	g := NewGomegaWithT(t)
	interceptor := ProducerInterceptor{
		ProviderId:       "pinger.services.cf.dev01.aws.istio.sapcloud.io",
		SystemDomain:     "services.cf.dev01.aws.istio.sapcloud.io",
		LoadBalancerPort: 9000,
		IpAddress:        "10.0.81.0",
		IstioDirectory:   os.TempDir(),
	}
	endpoints := []model.Endpoint{{"test.local", 5757}}
	bindResponse, err := interceptor.postBind(model.BindRequest{}, model.BindResponse{
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
	interceptor := ProducerInterceptor{
		ProviderId:       "pinger.services.cf.dev01.aws.istio.sapcloud.io",
		SystemDomain:     "services.cf.dev01.aws.istio.sapcloud.io",
		LoadBalancerPort: 9000,
		IpAddress:        "10.0.81.0",
		IstioDirectory:   os.TempDir(),
	}
	endpoints := []model.Endpoint{{"test.local", 5757}}
	_, err := interceptor.postBind(model.BindRequest{}, model.BindResponse{
		Credentials: model.Credentials{
			Endpoints: endpoints,
		},
	}, "123", adapt)
	g.Expect(err).NotTo(HaveOccurred())
	fileName := path.Join(interceptor.IstioDirectory, "123.yml")
	file, err := os.Open(fileName)
	g.Expect(err).NotTo(HaveOccurred())
	content, err := ioutil.ReadAll(file)
	g.Expect(err).NotTo(HaveOccurred())
	contentAsString := string(content)
	g.Expect(contentAsString).To(ContainSubstring("9000"))

	err = interceptor.postDelete("123")
	g.Expect(err).NotTo(HaveOccurred())
	_, err = os.Stat(fileName)
	g.Expect(err).To(HaveOccurred())

	err = interceptor.postDelete("123")
	g.Expect(err).NotTo(HaveOccurred())
}
