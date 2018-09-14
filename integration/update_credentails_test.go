// +build integration

package integration

import (
	"bytes"
	"crypto/tls"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	. "github.com/onsi/gomega"
	"net/http"
	"os"
	"path"
	"testing"
)

const baseUrl = "https://istio-broker.cfapps.dev01.aws.istio.sapcloud.io"

func TestAdaptCredentialsWithInvalidRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	body := []byte(`
                  {
                  "credentials": {
                   "dbname": "yLO2WoE0-mCcEppn",
                   "hostname": "10.11.241.0",
                   "password": "redacted",
                   "port": "47637",
                   "ports": {
                    "5432/tcp": "47637"
                   },
                   "uri": "postgres://mma4G8N0isoxe17v:redacted@10.11.241.0:47637/yLO2WoE0-mCcEppn",
                   "username": "mma4G8N0isoxe17v"
                  }
                  }`)
	client := http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	request, err := http.NewRequest(http.MethodPut, baseUrl+"/v2/service_instances/1/service_bindings/2/adapt_credentials", bytes.NewReader(body))
	g.Expect(err).NotTo(HaveOccurred())

	response, err := client.Do(request)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(response.StatusCode).To(Equal(400))

}

func TestAdaptCredentialsWithValidRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	body := []byte(`
                  {
                  "credentials": {
                   "dbname": "yLO2WoE0-mCcEppn",
                   "hostname": "10.11.241.0",
                   "password": "redacted",
                   "port": "47637",
                   "ports": {
                    "5432/tcp": "47637"
                   },
                   "uri": "postgres://mma4G8N0isoxe17v:redacted@10.11.241.0:47637/yLO2WoE0-mCcEppn",
                   "username": "mma4G8N0isoxe17v"
                  },
                  "endpoint_mappings": [{
                    "source": {"host": "10.11.241.0", "port": 47637},
                    "target": {"host": "appnethost", "port": 9876}
                  	}]
                  }`)
	client := http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	request, err := http.NewRequest(http.MethodPut, baseUrl+"/v2/service_instances/1/service_bindings/2/adapt_credentials", bytes.NewReader(body))
	g.Expect(err).NotTo(HaveOccurred())

	response, err := client.Do(request)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(response.StatusCode).To(Equal(200))

}

func TestAdaptCredentialsWithInvalidRequestUsingKubectl(t *testing.T) {

	const kubeBaseUrl = "http://broker:VoJniQuzmenuhsowelbahenhukejd755@istiobroker.catalog.svc.cluster.local:9999"

	g := NewGomegaWithT(t)

	kubeconfig := os.Getenv("KUBECONFIG")
	g.Expect(kubeconfig).NotTo(BeEmpty())

	kubectl := Kubectl{g}

	podName := kubectl.GetPod("istiobroker")

	result := put(g, kubectl, podName, kubeBaseUrl+"/v2/service_instances/1/service_bindings/2/adapt_credentials", `
	                  {
	                  "credentials": {
	                   "dbname": "yLO2WoE0-mCcEppn",
	                   "hostname": "10.11.241.0",
	                   "password": "redacted",
	                   "port": "47637",
	                   "ports": {
	                    "5432/tcp": "47637"
	                   },
	                   "uri": "postgres://mma4G8N0isoxe17v:redacted@10.11.241.0:47637/yLO2WoE0-mCcEppn",
	                   "username": "mma4G8N0isoxe17v"
	                  }
	                  }`)

	g.Expect(result).To(ContainSubstring(`HTTP/1.1 400 Bad Request`))
	g.Expect(result).To(ContainSubstring(`Invalid request`))
}

func TestAdaptCredentialsWithValidRequestUsingKubectl(t *testing.T) {

	g := NewGomegaWithT(t)

	kubeconfig := os.Getenv("KUBECONFIG")
	g.Expect(kubeconfig).NotTo(BeEmpty())

	kubectl := Kubectl{g}

	var clusterServiceBroker v1beta1.ClusterServiceBroker
	kubectl.Read(&clusterServiceBroker, "istiobroker")
	kubeBaseUrl := clusterServiceBroker.GetURL()

	podName := kubectl.GetPod("istiobroker")

	result := put(g, kubectl, podName, kubeBaseUrl+"/v2/service_instances/1/service_bindings/2/adapt_credentials", `{
                  "credentials": {
                   "dbname": "yLO2WoE0-mCcEppn",
                   "hostname": "10.11.241.0",
                   "password": "redacted",
                   "port": "47637",
                   "ports": {
                    "5432/tcp": "47637"
                   },
                   "uri": "postgres://mma4G8N0isoxe17v:redacted@10.11.241.0:47637/yLO2WoE0-mCcEppn",
                   "username": "mma4G8N0isoxe17v"
                  },
                  "endpoint_mappings": [{
                    "source": {"host": "10.11.241.0", "port": 47637},
                    "target": {"host": "new-magic-host", "port": 9876}
                  	}]
                  }`)

	g.Expect(result).To(ContainSubstring(`HTTP/1.1 200 OK`))
	g.Expect(result).To(ContainSubstring(`"hostname":"new-magic-host"`))
}

func put(g *GomegaWithT, kubectl Kubectl, podName string, url string, body string) string {
	const podFileName = "/tmp/post.json"
	fileName := path.Join(os.TempDir(), "post.json")
	file, err := os.Create(fileName)
	g.Expect(err).NotTo(HaveOccurred())
	file.Write([]byte(body))
	file.Close()
	kubectl.run("cp", fileName, "catalog/"+podName+":"+podFileName)
	os.Remove(fileName)
	result := kubectl.Exec(podName, "-n", "catalog", "-ti", "--", "curl",
		"-X", "PUT", "-v",
		"-H", escape("X-Broker-Api-Version: 2.12"),
		"-H", escape("Content-Type: application/json"),
		"-d", "@"+podFileName, url)
	return result
}

func escape(arg string) string {
	return "'" + arg + "'"
}
