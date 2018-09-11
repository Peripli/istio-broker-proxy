// +build integration

package integration

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	. "github.com/onsi/gomega"
	"os"
	"testing"
)

const service_instance = `apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: postgres-instance
spec:
  clusterServiceClassExternalName: postgresql
  clusterServicePlanExternalName: v9.4-dev`

const service_binding = `apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  name: postgres-binding
spec:
  instanceRef:
    name: postgres-instance`

func TestServiceBindingIsSuccessful(t *testing.T) {

	g := NewGomegaWithT(t)

	kubeconfig := os.Getenv("KUBECONFIG")
	g.Expect(kubeconfig).NotTo(BeEmpty())

	kubectl := Kubectl{g}

	// Test if list of available services is not empty
	var classes v1beta1.ClusterServiceClassList
	err := kubectl.List(&classes)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(classes.Items).NotTo(BeEmpty(), "List of available services in OSB should not be empty")

	// Clean up old services
	kubectl.Delete("ServiceBinding", "postgres-binding")
	kubectl.Delete("ServiceInstance", "postgres-instance")

	// Create service binding
	kubectl.Apply([]byte(service_instance))
	kubectl.Apply([]byte(service_binding))
	var status v1beta1.ServiceBinding
	err = kubectl.Get(&status, "postgres-binding")
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(status.Status.Conditions[0].Status).To(Equal(v1beta1.ConditionTrue))

}
