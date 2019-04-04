package router

import (
	. "github.com/onsi/gomega"
	"istio.io/istio/pilot/pkg/model"
	"testing"
)

func TestFileConfigStore(t *testing.T) {
	g := NewGomegaWithT(t)

	fileCS := fileConfigStore{}
	error := fileCS.CreateIstioConfig("binding-id", []model.Config{})

	g.Expect(error).NotTo(HaveOccurred())
}
