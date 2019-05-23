package integration

import (
	"encoding/json"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	. "github.com/onsi/gomega"
	"k8s.io/api/core/v1"
	"net/url"
	"strings"
	"testing"
)

const (
	service_instance_example_service = `---
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: integration-test-instance
spec:
  clusterServiceClassExternalName: example-service
  clusterServicePlanExternalName: plan-one`

 service_instance_no_istio_provider = `---
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: integration-test-instance
spec:
  clusterServiceClassExternalName: alternate-example-service
  clusterServicePlanExternalName: plan-one`

service_binding_example_service = `---
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  name: integration-test-binding
spec:
  instanceRef:
    name: integration-test-instance`

client_config_curl = `---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: client-curl
spec:
  template:
    metadata:
      labels:
        app: client-curl
    spec:
      containers:
      - image: gcr.io/peripli/curl
        imagePullPolicy: Always
        name: curl
`
)


type ServiceInstanceList struct {
	v1.ServiceList
}

func TestServiceBindingCanReachCF(t *testing.T) {
	skipWithoutKubeconfigSet(t)

	g := NewGomegaWithT(t)
	kubectl := NewKubeCtl(g)

	bindID := createServiceBinding(kubectl, g, "integration-test", service_instance_example_service, service_binding_example_service)
	podName := runClientPod(kubectl, client_config_curl, "client-curl")
	g.Expect(podName).To(ContainSubstring("client-curl"))

	var serviceSecret v1.Secret
	kubectl.Read(&serviceSecret, "integration-test-binding")

	url, err := url.Parse(string(serviceSecret.Data["url"]) + "payload")
	g.Expect(err).NotTo(HaveOccurred())

	serviceName := "svc-0-" + bindID + ".catalog.svc.cluster.local:5555"
	kubectl.Exec(podName, "-c", "curl", "-i", "--", "curl", "-o", "body.json", "-v", "--connect-to", url.Host + ":443:" + serviceName, url.String())

	bodyString := kubeFetchFile(kubectl, g, "body.json", podName, "curl")
	var body interface{}
	err = json.Unmarshal([]byte(bodyString), &body)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(body.(map[string]interface{})["body"]).To(Equal("payload"))
}

func TestServiceBindingIstioObjectsDeletedProperly(t *testing.T) {
	skipWithoutKubeconfigSet(t)

	g := NewGomegaWithT(t)
	kubectl := NewKubeCtl(g)

	bindID := createServiceBinding(kubectl, g, "integration-test", service_instance_example_service, service_binding_example_service)

	kubectl.Delete("ServiceBinding", "integration-test-binding")
	kubectl.Delete("ServiceInstance", "integration-test-instance")

	var serviceEntries ServiceEntryList
	kubectl.List(&serviceEntries, "--all-namespaces")
	matchingIstioObjectCount := 0
	for _, serviceEntry := range serviceEntries.Items {
		if strings.Contains(serviceEntry.Metadata.Name, bindID) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(0))

	var virtualServices VirtualServiceList
	kubectl.List(&virtualServices, "--all-namespaces")

	for _, virtualService := range virtualServices.Items {

		if strings.Contains(virtualService.Metadata.Name, bindID) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(0))

	var gateways GatewayList
	kubectl.List(&gateways, "--all-namespaces")
	matchingIstioObjectCount = 0

	for _, gateway := range gateways.Items {

		if strings.Contains(gateway.Metadata.Name, bindID) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(0))

	var destinationRules DestinationruleList
	kubectl.List(&destinationRules, "--all-namespaces")
	matchingIstioObjectCount = 0

	for _, destinationRule := range destinationRules.Items {

		if strings.Contains(destinationRule.Metadata.Name, bindID) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(0))

}

func TestServiceBindingWithNoMatchingIstioProvider(t *testing.T) {
	skipWithoutKubeconfigSet(t)

	g := NewGomegaWithT(t)
	kubectl := NewKubeCtl(g)

	createServiceBindingButNoIstioResources(kubectl, g, "integration-test", service_instance_no_istio_provider, service_binding_example_service)
}

func createServiceBindingButNoIstioResources(kubectl *kubectl, g *GomegaWithT, namePrefix string, serviceConfig string, bindingConfig string) string {
	// Test if list of available servicesInstance is not empty
	var classes v1beta1.ClusterServiceClassList
	kubectl.List(&classes)
	g.Expect(classes.Items).NotTo(BeEmpty(), "List of available servicesInstance in OSB should not be empty")
	kubectl.Delete("ServiceBinding", namePrefix+"-binding")
	kubectl.Delete("ServiceInstance", namePrefix+"-instance")
	kubectl.Apply([]byte(serviceConfig))
	waitForServiceInstance(kubectl, g, namePrefix)
	kubectl.Apply([]byte(bindingConfig))
	serviceBinding := waitForServiceBinding(kubectl, g, namePrefix)
	bindID := serviceBinding.Spec.ExternalID
	var servicesInstance ServiceInstanceList
	kubectl.List(&servicesInstance, "--all-namespaces=true")
	g.Expect(servicesInstance.Items).NotTo(BeEmpty(), "List of available servicesInstance in OSB should not be empty")
	g.Expect(serviceinstanceExists(servicesInstance, namePrefix)).To(BeTrue())
	var services v1.ServiceList
	kubectl.List(&services, "--all-namespaces=true")
	g.Expect(services.Items).NotTo(BeEmpty(), "List of available services in OSB should not be empty")
	g.Expect(serviceExists(services, bindID)).To(BeFalse())
	var serviceEntries ServiceEntryList
	kubectl.List(&serviceEntries, "--all-namespaces=true")
	matchingServiceEntryExists := false
	for _, serviceEntry := range serviceEntries.Items {
		if strings.Contains(serviceEntry.Metadata.Name, bindID) {
			matchingServiceEntryExists = true
		}
	}
	g.Expect(matchingServiceEntryExists).To(BeFalse())
	var virtualServices VirtualServiceList
	kubectl.List(&virtualServices, "--all-namespaces=true")
	matchingIstioObjectCount := 0
	for _, virtualService := range virtualServices.Items {

		if strings.Contains(virtualService.Metadata.Name, bindID) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(0))
	var gateways GatewayList
	kubectl.List(&gateways, "--all-namespaces=true")
	matchingIstioObjectCount = 0
	for _, gateway := range gateways.Items {

		if strings.Contains(gateway.Metadata.Name, bindID) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(0))
	var destinationRules DestinationruleList
	kubectl.List(&destinationRules, "--all-namespaces=true")
	matchingIstioObjectCount = 0
	for _, destinationRule := range destinationRules.Items {

		if strings.Contains(destinationRule.Metadata.Name, bindID) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(0))
	return bindID
}

func serviceinstanceExists(serviceInstaces ServiceInstanceList, name string) bool {
	for _, serviceInstance := range serviceInstaces.Items {
		if strings.Contains(serviceInstance.Name, name) {
			return true
		}
	}
	return false
}
