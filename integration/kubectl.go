// +build integration

package integration

import (
	"encoding/json"
	"fmt"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"k8s.io/api/core/v1"
	"os"
	"os/exec"
	"reflect"
	"strings"
)

func NewKubeCtl(g *GomegaWithT) *kubectl {
	kubeconfig := os.Getenv("KUBECONFIG")
	g.Expect(kubeconfig).NotTo(BeEmpty())

	return &kubectl{g}
}

type kubectl struct {
	g *GomegaWithT
}

func (self kubectl) run(args ...string) []byte {
	fmt.Println("kubectl ", strings.Join(args, " "))
	out, err := exec.Command("kubectl", args...).CombinedOutput()
	self.g.Expect(err).ShouldNot(HaveOccurred())
	// fmt.Println(string(out))
	return out
}

func (self kubectl) Delete(kind string, name string) {
	self.run("delete", kind, name, "--ignore-not-found=true")
}

func (self kubectl) Apply(fileBody []byte) {

	file, err := ioutil.TempFile("", "*")
	defer os.Remove(file.Name())

	self.g.Expect(err).ShouldNot(HaveOccurred())
	_, err = file.Write(fileBody)
	self.g.Expect(err).ShouldNot(HaveOccurred())
	err = file.Close()
	self.g.Expect(err).ShouldNot(HaveOccurred())
	self.run("apply", "-f", file.Name())
}

func (self kubectl) Read(result interface{}, name string) {
	kind := reflect.TypeOf(result).Elem().Name()
	response := self.run("get", kind, name, "-o", "json")
	err := json.Unmarshal(response, result)
	self.g.Expect(err).ShouldNot(HaveOccurred())
}

func (self kubectl) Exec(podName string, args ...string) string {
	cmd := append([]string{"exec", podName}, args...)
	return string(self.run(cmd...))
}

func (self kubectl) List(result interface{}, args ...string) {
	kind := reflect.TypeOf(result).Elem().Name()
	if strings.HasSuffix(kind, "List") {
		kind = kind[0 : len(kind)-4]
	}
	args = append([]string{"get", kind, "-o", "json"}, args...)
	response := self.run(args...)
	err := json.Unmarshal(response, result)
	self.g.Expect(err).ShouldNot(HaveOccurred())
}

func (self kubectl) GetPod(args ...string) string {
	var pods v1.PodList
	self.List(&pods, args...)
	self.g.Expect(pods.Items).To(HaveLen(1), "Pod not found")
	podName := pods.Items[0].Name
	return podName
}
