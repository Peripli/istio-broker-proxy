package main

import (
	"flag"
	"github.com/Peripli/istio-broker-proxy/pkg/router"
	. "github.com/onsi/gomega"
	"os"
	"strings"
	"testing"
)

func TestSetupConfiguration(t *testing.T) {
	g := NewGomegaWithT(t)
	args := strings.Split("--port 8000 --forwardUrl https://192.168.252.10:9293/cf --skipVerifyTLS --systemdomain istio.xxx.io --providerId istio.yyy.io --loadBalancerPort 9000 --istioDirectory /var/vcap/store/istio-config --ipAddress 10.0.81.0 --serviceNamePrefix istio- ", " ")
	flag.CommandLine.Parse(args)

	g.Expect(routerConfig.SkipVerifyTLS).To(BeTrue())
	g.Expect(producerInterceptor.ProviderId).To(Equal("istio.yyy.io"))
}

func TestConfigureProducerInterceptor(t *testing.T) {
	g := NewGomegaWithT(t)

	args := strings.Split("--port 8000 --forwardUrl https://192.168.252.10:9293/cf --skipVerifyTLS --systemdomain istio.xxx.io --providerId istio.yyy.io --loadBalancerPort 9000 --istioDirectory /tmp --ipAddress 10.0.81.0 --serviceNamePrefix istio- --networkProfile xxx.yyy", " ")
	flag.CommandLine.Parse(args)
	interceptor, ok := configureInterceptor(newMockConfigStore).(router.ProducerInterceptor)

	g.Expect(ok).To(BeTrue())

	g.Expect(interceptor.NetworkProfile).To(Equal("xxx.yyy"))
	g.Expect(interceptor.SystemDomain).To(Equal("istio.xxx.io"))
}

func TestConfigureConsumerInterceptor(t *testing.T) {
	g := NewGomegaWithT(t)

	args := strings.Split("--port 8000 --forwardUrl https://192.168.252.10:9293/cf --skipVerifyTLS --consumerId istio.yyy.io --loadBalancerPort 9000 --istioDirectory /tmp --ipAddress 10.0.81.0 --serviceNamePrefix istio- --networkProfile xxx.yyy", " ")
	flag.CommandLine.Parse(args)
	interceptor, ok := configureInterceptor(newMockConfigStore).(router.ConsumerInterceptor)

	g.Expect(ok).To(BeTrue())

	g.Expect(interceptor.NetworkProfile).To(Equal("xxx.yyy"))
	g.Expect(interceptor.ConsumerId).To(Equal("istio.yyy.io"))
}

func newMockConfigStore() router.ConfigStore {
	return &router.MockConfigStore{}
}

func TestMain(m *testing.M) {
	SetupConfiguration()
	os.Exit(m.Run())
}
