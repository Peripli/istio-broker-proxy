package integration

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	. "github.com/onsi/gomega"
	"os"
	"path"
	"testing"
)

func TestAdaptCredentialsWithInvalidRequest(t *testing.T) {
	skipWithoutKubeconfigSet(t)

	g := NewGomegaWithT(t)
	kubectl := NewKubeCtl(g)

	kubeBaseUrl := getClusterServiceBrokerUrl(kubectl)
	podName := kubectl.GetPod("-n", "catalog", "-l", "app=istiobroker")

	result := post(g, kubectl, podName, kubeBaseUrl+"/v2/service_instances/1/service_bindings/2/adapt_credentials", `
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
	g.Expect(result).To(ContainSubstring(`No endpoint mappings available`))
}

func TestAdaptCredentialsWithValidRequest(t *testing.T) {
	skipWithoutKubeconfigSet(t)

	g := NewGomegaWithT(t)
	kubectl := NewKubeCtl(g)

	kubeBaseUrl := getClusterServiceBrokerUrl(kubectl)

	podName := kubectl.GetPod("-n", "catalog", "-l", "app=istiobroker")

	result := post(g, kubectl, podName, kubeBaseUrl+"/v2/service_instances/1/service_bindings/2/adapt_credentials", `{
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

func getClusterServiceBrokerUrl(kubectl *kubectl) string {
	var clusterServiceBroker v1beta1.ClusterServiceBroker
	kubectl.Read(&clusterServiceBroker, "istiobroker")
	kubeBaseUrl := clusterServiceBroker.GetURL()
	return kubeBaseUrl
}

func post(g *GomegaWithT, kubectl *kubectl, podName string, url string, body string) string {
	const podFileName = "/tmp/post.json"
	fileName := path.Join(os.TempDir(), "post.json")
	file, err := os.Create(fileName)
	g.Expect(err).NotTo(HaveOccurred())
	file.Write([]byte(body))
	file.Close()
	kubectl.run("cp", fileName, "catalog/"+podName+":"+podFileName)
	os.Remove(fileName)
	result := kubectl.Exec(podName, "-n", "catalog", "-ti", "--", "curl",
		"-X", "POST", "-v",
		"-H", escape("X-Broker-Api-Version: 2.12"),
		"-H", escape("Content-Type: application/json"),
		"-d", "@"+podFileName, url)
	return result
}

func escape(arg string) string {
	return "'" + arg + "'"
}
