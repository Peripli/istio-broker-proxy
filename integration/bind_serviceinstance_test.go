package integration

import (
	"fmt"
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

//FIXME There is much code duplicated (waiter)
func createServiceBindingButNoIstioResources(kubectl *kubectl, g *GomegaWithT, namePrefix string, serviceConfig string, bindingConfig string) string {
	// Test if list of available servicesInstance is not empty
	var classes v1beta1.ClusterServiceClassList
	kubectl.List(&classes)
	g.Expect(classes.Items).NotTo(BeEmpty(), "List of available servicesInstance in OSB should not be empty")
	kubectl.Delete("ServiceBinding", namePrefix+"-binding")
	kubectl.Delete("ServiceInstance", namePrefix+"-instance")
	kubectl.Apply([]byte(serviceConfig))
	var serviceInstance v1beta1.ServiceInstance
	waitForCompletion(g, func() bool {
		kubectl.Read(&serviceInstance, namePrefix+"-instance")
		statusLen := len(serviceInstance.Status.Conditions)
		if statusLen == 0 {
			return false
		}

		condition := serviceInstance.Status.Conditions[statusLen-1]

		if condition.Status != v1beta1.ConditionTrue {
			return false
		}

		g.Expect(condition.Type).To(Equal(v1beta1.ServiceInstanceConditionReady))
		return true
	}, "serviceinstance")
	kubectl.Apply([]byte(bindingConfig))
	var serviceBinding v1beta1.ServiceBinding
	waitForCompletion(g, func() bool {
		kubectl.Read(&serviceBinding, namePrefix+"-binding")
		statusLen := len(serviceBinding.Status.Conditions)
		if statusLen == 0 {
			return false
		}

		condition := serviceBinding.Status.Conditions[statusLen-1]
		if condition.Status != v1beta1.ConditionTrue {
			return false
		}

		g.Expect(condition.Type).To(Equal(v1beta1.ServiceBindingConditionReady), fmt.Sprintf("Is not ready: %s", string(condition.Reason)))
		return true
	}, "servicebinding")
	bindId := serviceBinding.Spec.ExternalID
	var servicesInstance ServiceInstanceList
	kubectl.List(&servicesInstance, "--all-namespaces=true")
	g.Expect(servicesInstance.Items).NotTo(BeEmpty(), "List of available servicesInstance in OSB should not be empty")
	g.Expect(serviceinstanceExists(servicesInstance, namePrefix)).To(BeTrue())
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
