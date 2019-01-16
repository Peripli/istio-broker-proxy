package main

import (
	"flag"
	. "github.com/onsi/gomega"
	"strings"
	"testing"
)

func TestSetupConfiguration(t *testing.T) {
	g := NewGomegaWithT(t)

	args := strings.Split("--port 8000 --forwardUrl https://192.168.252.10:9293/cf --skipVerifyTLS --systemdomain istio.xxx.io --providerId istio.yyy.io --loadBalancerPort 9000 --istioDirectory /var/vcap/store/istio-config --ipAddress 10.0.81.0 --serviceNamePrefix istio- ", " ")
	SetupConfiguration()
	flag.CommandLine.Parse(args)

	g.Expect(routerConfig.SkipVerifyTLS).To(BeTrue())
	g.Expect(producerInterceptor.ProviderId).To(Equal("istio.yyy.io"))
}
