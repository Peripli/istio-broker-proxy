// +build integration

package integration

import (
	"encoding/json"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"
)

type Kubectl struct {
	g *GomegaWithT
}

func (self Kubectl) run(args ...string) []byte {
	log.Println("kubectl ", strings.Join(args, " "))
	out, err := exec.Command("kubectl", args...).Output()
	self.g.Expect(err).ShouldNot(HaveOccurred())
	// fmt.Println(string(out))
	return out
}

func (self Kubectl) Delete(kind string, name string) {
	self.run("delete", kind, name, "--ignore-not-found=true")
}

func (self Kubectl) Apply(fileBody []byte) {

	file, err := ioutil.TempFile("", "*")
	defer os.Remove(file.Name())

	self.g.Expect(err).ShouldNot(HaveOccurred())
	_, err = file.Write(fileBody)
	self.g.Expect(err).ShouldNot(HaveOccurred())
	err = file.Close()
	self.g.Expect(err).ShouldNot(HaveOccurred())
	self.run("apply", "-f", file.Name())
}

func (self Kubectl) Get(result interface{}, name string) error {
	kind := reflect.TypeOf(result).Elem().Name()
	response := self.run("get", kind, name, "-o", "json")
	return json.Unmarshal(response, result)
}

func (self Kubectl) List(result interface{}) error {
	kind := reflect.TypeOf(result).Elem().Name()
	if strings.HasSuffix(kind, "List") {
		kind = kind[0 : len(kind)-4]
	}
	response := self.run("get", kind, "-o", "json")
	return json.Unmarshal(response, result)
}
