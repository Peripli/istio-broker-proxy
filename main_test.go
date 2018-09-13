package main

import (
	. "github.com/onsi/gomega"
	"os"
	"testing"
)

func TestDefaultPortUsed(t *testing.T) {
	g := NewGomegaWithT(t)

	readPort()

	g.Expect(port).To(Equal(DefaultPort))
}

func TestCustomPortUsed(t *testing.T) {
	g := NewGomegaWithT(t)
	oldPort := os.Getenv("PORT")
	defer func() {
		os.Setenv("PORT", oldPort)
	}()
	expectedPort := "1234"
	os.Setenv("PORT", expectedPort)

	readPort()

	g.Expect(port).To(Equal(1234))
}
