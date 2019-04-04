package router

import (
	. "github.com/onsi/gomega"
	"io/ioutil"
	"istio.io/istio/pilot/pkg/model"
	"k8s.io/api/core/v1"
	"os"
	"path"
	"testing"
)

func newTmpFileConfigStore() ConfigStore {
	return NewFileConfigStore(os.TempDir())
}

func TestFileConfigStore(t *testing.T) {
	g := NewGomegaWithT(t)

	fileCS := newTmpFileConfigStore()
	err := fileCS.CreateIstioConfig("binding-id", []model.Config{})

	g.Expect(err).NotTo(HaveOccurred())
}

func TestFileConfigStoreInvalidDirectory(t *testing.T) {
	g := NewGomegaWithT(t)

	fileCS := &fileConfigStore{istioDirectory: "/invalid-directory"}
	err := fileCS.CreateIstioConfig("binding-id", []model.Config{})

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("unable to write istio configuration to file"))
}

func TestFileConfigStoreDeleteBinding(t *testing.T) {
	g := NewGomegaWithT(t)
	dir := os.TempDir()
	fileCS := NewFileConfigStore(dir)

	err := ioutil.WriteFile(path.Join(dir, "binding-id.yml"), []byte("hello\ngo\n"), 0644)
	g.Expect(err).NotTo(HaveOccurred())
	err = fileCS.DeleteBinding("binding-id")

	g.Expect(err).NotTo(HaveOccurred())
}

func TestFileConfigStoreDeleteBindingNotFound(t *testing.T) {
	g := NewGomegaWithT(t)

	fileCS := newTmpFileConfigStore()
	err := fileCS.DeleteBinding("binding-id")

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("Error during removal of file"))
}

func TestFileConfigStoreCreateService(t *testing.T) {
	g := NewGomegaWithT(t)

	fileCS := newTmpFileConfigStore()
	_, err := fileCS.CreateService("binding-id", &v1.Service{})

	g.Expect(err).To(HaveOccurred())
}
