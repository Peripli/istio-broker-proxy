package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

func TestCmdClientClientCallDiffers(t *testing.T) {
	g := NewGomegaWithT(t)
	stdoutServer := callClient("myservice", "myhost", 1234, "1.2.3.4", "5.4.3.1")

	g.Expect(stdoutServer).NotTo(BeEmpty())
}

func callClient(serviceName string, hostVirtualService string, portServiceEntry int, endpointServiceEntry string, serviceIp string) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	createOutput(serviceName, hostVirtualService, portServiceEntry, endpointServiceEntry, serviceIp)
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
