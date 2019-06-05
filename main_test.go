package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/Peripli/istio-broker-proxy/pkg/router"
	. "github.com/onsi/gomega"
	"istio.io/istio/pkg/log"
	"os"
	"os/exec"
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

func TestLogLevelDebug(t *testing.T) {
	g := NewGomegaWithT(t)
	fmt.Printf("%s", os.Getenv("LOG"))
	if os.Getenv("LOG") == "1" {
		args := strings.Split("--logLevel 5", " ")
		flag.CommandLine.Parse(args)
		configureLogging()
		log.Debugf("Debug logging enabled")
		log.Sync()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestLogLevelDebug")
	cmd.Env = append(os.Environ(), "LOG=1")
	var out, err bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &err
	cmd.Run()
	fmt.Println(out.String())
	fmt.Println(err.String())
	g.Expect(out.String()).To(ContainSubstring("Log level is set to 5"))
	g.Expect(out.String()).To(ContainSubstring("Debug logging enabled"))
}

func TestDefaultLogLevelInfo(t *testing.T) {
	g := NewGomegaWithT(t)
	fmt.Printf("%s", os.Getenv("LOG"))
	if os.Getenv("LOG") == "1" {
		flag.CommandLine.Parse([]string{})
		configureLogging()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestDefaultLogLevelInfo")
	cmd.Env = append(os.Environ(), "LOG=1")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()
	g.Expect(out.String()).To(ContainSubstring("Log level is set to 4"))
}

func TestNoDebugMessageInInfo(t *testing.T) {
	g := NewGomegaWithT(t)
	fmt.Printf("%s", os.Getenv("LOG"))
	if os.Getenv("LOG") == "1" {
		args := strings.Split("--logLevel 4 --networkProfile xxx.yyy", " ")
		flag.CommandLine.Parse(args)
		configureLogging()
		log.Debugf("Debug logging enabled")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestDefaultLogLevelInfo")
	cmd.Env = append(os.Environ(), "LOG=1")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()
	g.Expect(out.String()).To(ContainSubstring("Log level is set to 4"))
	g.Expect(out.String()).NotTo(ContainSubstring("Debug logging enabled"))
}
