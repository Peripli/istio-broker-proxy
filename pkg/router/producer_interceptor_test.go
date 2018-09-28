package router

import (
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestDefaultConfigurationIsWritten(t *testing.T) {
	g := NewGomegaWithT(t)
	NewProducerInterceptor(ProducerConfig{ProviderId: "your-provider", SystemDomain: "services.domain"}, 147)
	file, err := os.Open(path.Join(os.TempDir(), "istio-broker.yml"))
	g.Expect(err).NotTo(HaveOccurred())
	content, err := ioutil.ReadAll(file)
	contentAsString := string(content)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(contentAsString).To(ContainSubstring("147"))
	g.Expect(contentAsString).To(ContainSubstring("istio-broker.services.domain"))
	g.Expect(contentAsString).To(MatchRegexp("number: 9000"))

}

func TestYmlFileIsCorrectlyWritten(t *testing.T) {
	g := NewGomegaWithT(t)
	///var/vcap/packages/istio-broker/bin/istio-broker --port 8000 --forwardUrl https://10.11.252.10:9293/cf
	// --systemdomain services.cf.dev01.aws.istio.sapcloud.io --ProviderId pinger.services.cf.dev01.aws.istio.sapcloud.io
	// --LoadBalancerPort 9000 --istioDirectory /var/vcap/store/istio-config --ipAddress 10.0.81.0
	NewProducerInterceptor(ProducerConfig{
		ProviderId:       "pinger.services.cf.dev01.aws.istio.sapcloud.io",
		SystemDomain:     "services.cf.dev01.aws.istio.sapcloud.io",
		LoadBalancerPort: 9000,
		IpAddress:        "10.0.81.0",
	}, 8000)

	file, err := os.Open(path.Join(os.TempDir(), "istio-broker.yml"))
	g.Expect(err).NotTo(HaveOccurred())
	content, err := ioutil.ReadAll(file)
	contentAsString := string(content)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(contentAsString).To(ContainSubstring("8000"))
	g.Expect(contentAsString).To(ContainSubstring("istio-broker.services.cf.dev01.aws.istio.sapcloud.io"))
	g.Expect(contentAsString).To(MatchRegexp("number: 9000"))

}
