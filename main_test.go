package main

import (
	"flag"
	"github.com/Peripli/istio-broker-proxy/pkg/router"
	. "github.com/onsi/gomega"
	"os"
	"strings"
	"testing"
)


func newMockConfigStore(dummy string) router.ConfigStore {
	return router.NewMockConfigStore()
}

func TestSetupConfiguration(t *testing.T) {
	g := NewGomegaWithT(t)
	args := strings.Split("--port 8000 --forwardUrl https://192.168.252.10:9293/cf --skipVerifyTLS --systemdomain istio.xxx.io --providerId istio.yyy.io --loadBalancerPort 9000 --configStore file:///var/vcap/store/istio-config --ipAddress 10.0.81.0 --serviceNamePrefix istio- ", " ")
	flag.CommandLine.Parse(args)

	g.Expect(routerConfig.SkipVerifyTLS).To(BeTrue())
	g.Expect(producerInterceptor.ProviderID).To(Equal("istio.yyy.io"))
}

func TestConfigureProducerInterceptor(t *testing.T) {
	g := NewGomegaWithT(t)

	args := strings.Split("--port 8000 --forwardUrl https://192.168.252.10:9293/cf --skipVerifyTLS --systemdomain istio.xxx.io --providerId istio.yyy.io --loadBalancerPort 9000 --configStore file:///tmp --ipAddress 10.0.81.0 --serviceNamePrefix istio- --networkProfile xxx.yyy", " ")
	flag.CommandLine.Parse(args)
	interceptor, ok := configureInterceptor(newMockConfigStore).(router.ProducerInterceptor)

	g.Expect(ok).To(BeTrue())

	g.Expect(interceptor.NetworkProfile).To(Equal("xxx.yyy"))
	g.Expect(interceptor.SystemDomain).To(Equal("istio.xxx.io"))
}

func TestConfigureConsumerInterceptor(t *testing.T) {
	g := NewGomegaWithT(t)

	args := strings.Split("--port 8000 --forwardUrl https://192.168.252.10:9293/cf --skipVerifyTLS --consumerId istio.yyy.io --loadBalancerPort 9000 --configStore file:///tmp --ipAddress 10.0.81.0 --serviceNamePrefix istio- --networkProfile xxx.yyy", " ")
	flag.CommandLine.Parse(args)
	interceptor, ok := configureInterceptor(newMockConfigStore).(router.ConsumerInterceptor)

	g.Expect(ok).To(BeTrue())

	g.Expect(interceptor.NetworkProfile).To(Equal("xxx.yyy"))
	g.Expect(interceptor.ConsumerID).To(Equal("istio.yyy.io"))
}

func TestMain(m *testing.M) {
	SetupConfiguration()
	os.Exit(m.Run())
}

func TestNewConfigStoreInvalidSchema(t *testing.T) {
	g := NewGomegaWithT(t)
	_, err := newConfigStore("xxx://")
	g.Expect(err).To(HaveOccurred())
}

func TestNewConfigStoreFile(t *testing.T) {
	g := NewGomegaWithT(t)
	_, err := newConfigStore("file:///tmp")
	g.Expect(err).NotTo(HaveOccurred())
}

func TestNewConfigStoreInvalidURL(t *testing.T) {
	g := NewGomegaWithT(t)
	_, err := newConfigStore("\x7f")
	g.Expect(err).To(HaveOccurred())
}
