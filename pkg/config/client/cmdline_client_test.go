package main

import (
	"bytes"
	"github.com/Peripli/istio-broker-proxy/pkg/router"
	"io"
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

func TestCmdClientClientCallDiffers(t *testing.T) {
	g := NewGomegaWithT(t)
	stdoutServer := callClient( false,"myservice", "myhost", 1234, "1.2.3.4", "my.arbitrary.domain.io", false, nil)

	g.Expect(stdoutServer).NotTo(BeEmpty())
}

func TestCmdClientClientCreatesService(t *testing.T) {
	g := NewGomegaWithT(t)
	mockConfigStore := router.MockConfigStore{}
	callClient( true,"myservice", "myhost", 1234, "", "my.arbitrary.domain.io", false, &mockConfigStore)

	g.Expect(mockConfigStore.CreatedServices).To(HaveLen(1))
	createdService := mockConfigStore.CreatedServices[0]
	g.Expect(createdService.ObjectMeta.Labels["istio-broker-proxy-binding-id"]).To(Equal("client-binding-myservice"))
}

func TestCmdClientClientDeletesService(t *testing.T) {
	g := NewGomegaWithT(t)
	mockConfigStore := router.MockConfigStore{}
	callClient( true,"myservice", "myhost", 1234, "", "my.arbitrary.domain.io", false, &mockConfigStore)
	callClient( true,"myservice", "myhost", 1234, "", "my.arbitrary.domain.io", true, &mockConfigStore)

	g.Expect(mockConfigStore.DeletedServices).To(HaveLen(1))
	deletedService := mockConfigStore.DeletedServices[0]
	g.Expect(deletedService).To(Equal("myservice"))
}

func callClient(clientConfig bool, serviceName string, hostVirtualService string, portServiceEntry int, endpointServiceEntry string, systemDomain string, delete bool, configStore router.ConfigStore) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	createOutput(clientConfig, serviceName, hostVirtualService, portServiceEntry, endpointServiceEntry, systemDomain, delete, configStore)
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()
	w.Close()
	os.Stdout = oldStdout
	stdout := <-outC
	return stdout
}
