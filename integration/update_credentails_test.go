package integration

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	. "github.com/onsi/gomega"
	"os"
	"path"
)

func getPostgresqlServiceBroker(kubectl *kubectl) string {
	var serviceClasses ClusterServiceClassList
	kubectl.List(&serviceClasses)
	for _, serviceClass := range serviceClasses.Items {
		if serviceClass.Spec.ExternalName == "postgresql" {
			return serviceClass.Spec.ClusterServiceBrokerName
		}
	}

	panic("No broker for postgresql registered.")
}

func getClusterServiceBrokerUrl(kubectl *kubectl) string {
	serviceBrokerName := getPostgresqlServiceBroker(kubectl)

	var clusterServiceBroker v1beta1.ClusterServiceBroker
	kubectl.Read(&clusterServiceBroker, serviceBrokerName)
	kubeBaseUrl := clusterServiceBroker.GetURL()
	return kubeBaseUrl
}

func post(g *GomegaWithT, kubectl *kubectl, namespace string, podName string, url string, body string) string {
	const podFileName = "/tmp/post.json"
	fileName := path.Join(os.TempDir(), "post.json")
	file, err := os.Create(fileName)
	g.Expect(err).NotTo(HaveOccurred())
	file.Write([]byte(body))
	file.Close()
	kubectl.run("cp", fileName, namespace+"/"+podName+":"+podFileName)
	os.Remove(fileName)
	result := kubectl.Exec(podName, "-n", namespace, "-ti", "--", "curl",
		"-X", "POST", "-v",
		"-H", escape("X-Broker-Api-Version: 2.12"),
		"-H", escape("Content-Type: application/json"),
		"-d", "@"+podFileName, url)
	return result
}

func escape(arg string) string {
	return "'" + arg + "'"
}
