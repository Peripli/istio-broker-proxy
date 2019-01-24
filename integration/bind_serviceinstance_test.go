package integration

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	. "github.com/onsi/gomega"
	"k8s.io/api/core/v1"
	"strings"
	"testing"
)

const service_instance_no_istio_provider = `---
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: integration-test-instance
spec:
  clusterServiceClassExternalName: example-service-integration-test
  clusterServicePlanExternalName: integration-test-plan-one`

const service_instance_config_no_istio_provider = `---
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  name: integration-test-binding
spec:
  instanceRef:
    name: integration-test-instance`

type ServiceInstanceList struct {
	v1.ServiceList
}

func TestServiceBindingWithNoMatchingIstioProvider(t *testing.T) {
	skipWithoutKubeconfigSet(t)

	g := NewGomegaWithT(t)
	kubectl := NewKubeCtl(g)

	createServiceBindingButNoIstioResources(kubectl, g, "integration-test", service_instance_no_istio_provider, service_instance_config_no_istio_provider)
}

func createServiceBindingButNoIstioResources(kubectl *kubectl, g *GomegaWithT, name string, serviceConfig string, bindingConfig string) string {
	// Test if list of available servicesInstance is not empty
	var classes v1beta1.ClusterServiceClassList
	kubectl.List(&classes)
	g.Expect(classes.Items).NotTo(BeEmpty(), "List of available servicesInstance in OSB should not be empty")
	kubectl.Delete("ServiceBinding", name+"-binding")
	kubectl.Delete("ServiceInstance", name+"-instance")
	kubectl.Apply([]byte(serviceConfig))
	var serviceInstance v1beta1.ServiceInstance
	waitForCompletion(g, func() bool {
		kubectl.Read(&serviceInstance, name+"-instance")
		statusLen := len(serviceInstance.Status.Conditions)
		if statusLen == 0 {
			return false
		}

		if serviceInstance.Status.Conditions[statusLen-1].Status != v1beta1.ConditionTrue {
			return false
		}

		g.Expect(serviceInstance.Status.Conditions[statusLen-1].Type).To(Equal(v1beta1.ServiceInstanceConditionReady))
		g.Expect(serviceInstance.Status.Conditions[statusLen-1].Status).To(Equal(v1beta1.ConditionTrue))
		return true
	})
	kubectl.Apply([]byte(bindingConfig))
	var serviceBinding v1beta1.ServiceBinding
	waitForCompletion(g, func() bool {
		kubectl.Read(&serviceBinding, name+"-binding")
		statusLen := len(serviceBinding.Status.Conditions)
		if statusLen == 0 {
			return false
		}

		if serviceBinding.Status.Conditions[statusLen-1].Status != v1beta1.ConditionTrue {
			return false
		}

		g.Expect(serviceBinding.Status.Conditions[statusLen-1].Type).To(Equal(v1beta1.ServiceBindingConditionReady))
		g.Expect(serviceBinding.Status.Conditions[statusLen-1].Status).To(Equal(v1beta1.ConditionTrue))
		return true
	})
	bindId := serviceBinding.Spec.ExternalID
	var servicesInstance ServiceInstanceList
	kubectl.List(&servicesInstance, "--all-namespaces=true")
	g.Expect(servicesInstance.Items).NotTo(BeEmpty(), "List of available servicesInstance in OSB should not be empty")
	g.Expect(serviceinstanceExists(servicesInstance, name)).To(BeTrue())
	var services v1.ServiceList
	kubectl.List(&services, "--all-namespaces=true")
	g.Expect(services.Items).NotTo(BeEmpty(), "List of available services in OSB should not be empty")
	g.Expect(serviceExists(services, bindId)).To(BeFalse())
	var serviceEntries ServiceEntryList
	kubectl.List(&serviceEntries, "--all-namespaces=true")
	matchingServiceEntryExists := false
	for _, serviceEntry := range serviceEntries.Items {
		if strings.Contains(serviceEntry.Metadata.Name, bindId) {
			matchingServiceEntryExists = true
		}
	}
	g.Expect(matchingServiceEntryExists).To(BeFalse())
	var virtualServices VirtualServiceList
	kubectl.List(&virtualServices, "--all-namespaces=true")
	matchingIstioObjectCount := 0
	for _, virtualService := range virtualServices.Items {

		if strings.Contains(virtualService.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(0))
	var gateways GatewayList
	kubectl.List(&gateways, "--all-namespaces=true")
	matchingIstioObjectCount = 0
	for _, gateway := range gateways.Items {

		if strings.Contains(gateway.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(0))
	var destinationRules DestinationruleList
	kubectl.List(&destinationRules, "--all-namespaces=true")
	matchingIstioObjectCount = 0
	for _, destinationRule := range destinationRules.Items {

		if strings.Contains(destinationRule.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(0))
	return bindId
}

func serviceinstanceExists(serviceInstaces ServiceInstanceList, name string) bool {
	for _, serviceInstance := range serviceInstaces.Items {
		if strings.Contains(serviceInstance.Name, name) {
			return true
		}
	}
	return false
}
