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
	stdoutClient := callClient(true, "myservice", "myhost", 1234, "1.2.3.4")
	stdoutServer := callClient(false, "myservice", "myhost", 1234, "1.2.3.4")

	g.Expect(stdoutClient).NotTo(Equal(stdoutServer))
}

func callClient(clientConfig bool, serviceName string, hostVirtualService string, portServiceEntry int, endpointServiceEntry string) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	createOutput(clientConfig, serviceName, hostVirtualService, portServiceEntry, endpointServiceEntry)
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
