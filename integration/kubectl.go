package integration

import (
	"encoding/json"
	"fmt"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"k8s.io/api/core/v1"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"
)

const MAX_WAITING_TIME = time.Duration(3) * time.Minute // delete namespace might take long
const ITERATION_WAITING_TIME = time.Duration(5) * time.Second

func NewKubeCtl(g *GomegaWithT) *kubectl {
	kubeconfig := os.Getenv("KUBECONFIG")
	g.Expect(kubeconfig).NotTo(BeEmpty(), "KUBECONFIG not set")

	return &kubectl{g}
}

type kubectl struct {
	g *GomegaWithT
}

func (self kubectl) run(args ...string) []byte {
	start := time.Now()
	expiry := time.Now().Add(MAX_WAITING_TIME)
	for {
		self.g.Expect(time.Now().Before(expiry)).To(BeTrue(),
			fmt.Sprintf("Timeout expired."))

		log.Println("kubectl ", strings.Join(args, " "))
		out, err := exec.Command("kubectl", args...).CombinedOutput()
		if err == nil {
			return out
		}

		if strings.Contains(err.Error(), "ServiceUnavailable") {
			expiry = start.Add(time.Duration(15) * time.Minute)
		}

		log.Printf("retry because of error: %s(%s) ", string(out), err.Error())

		time.Sleep(ITERATION_WAITING_TIME)
	}
}

func (self kubectl) runTailingOutput(args ...string) {
	log.Println("kubectl ", strings.Join(args, " "))
	command := exec.Command("kubectl", args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err := command.Run()
	self.g.Expect(err).NotTo(HaveOccurred(), "Error running kubectl")
}

func (self kubectl) CreateNamespace(name string) {
	self.run("create", "namespace", name)
}

func (self kubectl) Delete(kind string, name string) {
	self.run("delete", kind, name, "--ignore-not-found=true")
}

func (self kubectl) DeleteWithNamespace(kind string, name string, namespace string) {
	self.run("delete", kind, name, "-n", namespace, "--ignore-not-found=true")
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

func (self kubectl) Describe(result interface{}, name string) string {
	kind := reflect.TypeOf(result).Elem().Name()
	response := self.run("describe", kind, name)
	return string(response)
}

func (self kubectl) Exec(podName string, args ...string) {
	cmd := append([]string{"exec", podName}, args...)
	self.runTailingOutput(cmd...)
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
	length := len(pods.Items)
	self.g.Expect(length).Should(BeNumerically(">=", 1), "Pod not found")
	podName := pods.Items[0].Name
	return podName
}

func (self kubectl) GetPodIfExists(args ...string) string {
	var pods v1.PodList
	self.List(&pods, args...)
	if len(pods.Items) == 0 {
		return ""
	}
	podName := pods.Items[0].Name
	return podName
}

func (self kubectl) RolloutStatus(appName string) {
	self.runTailingOutput("rollout", "status", "deployment.v1.apps/"+appName)
}
