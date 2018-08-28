package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

func TestCmdClientWritesToStdout(t *testing.T) {
	g := NewGomegaWithT(t)
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	var args []string
	args = append(args, "--help")
	os.Args = args
	main()

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	os.Stdout = oldStdout
	stdout := <-outC

	g.Expect(stdout).NotTo(BeEmpty())
	fmt.Print(stdout)
}
