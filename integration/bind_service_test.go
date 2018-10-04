package integration

import (
	"fmt"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	. "github.com/onsi/gomega"
	"os"
	"testing"
	"time"
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

func skipWithoutKubeconfigSet(t *testing.T) {
	if os.Getenv("KUBECONFIG") == "" {
		t.Skip("KUBECONFIG not set, skipping integration test.")
	}
}

func TestServiceBindingIsSuccessful(t *testing.T) {
	skipWithoutKubeconfigSet(t)

	g := NewGomegaWithT(t)
	kubectl := NewKubeCtl(g)

	// Test if list of available services is not empty
	var classes v1beta1.ClusterServiceClassList
	kubectl.List(&classes)
	g.Expect(classes.Items).NotTo(BeEmpty(), "List of available services in OSB should not be empty")

	// Clean up old services
	kubectl.Delete("ServiceBinding", "postgres-binding")
	kubectl.Delete("ServiceInstance", "postgres-instance")

	// Create service binding
	kubectl.Apply([]byte(service_instance))
	var serviceInstance v1beta1.ServiceInstance
	waitForCompletion(g, func() bool {
		kubectl.Read(&serviceInstance, "postgres-instance")
		if len(serviceInstance.Status.Conditions) == 0 || serviceInstance.Status.Conditions[0].Status == v1beta1.ConditionUnknown {
			return false
		}
		g.Expect(serviceInstance.Status.Conditions[0].Status).To(Equal(v1beta1.ConditionTrue))
		return true
	})
	kubectl.Apply([]byte(service_binding))
	var serviceBinding v1beta1.ServiceBinding
	waitForCompletion(g, func() bool {
		kubectl.Read(&serviceBinding, "postgres-binding")
		if len(serviceBinding.Status.Conditions) == 0 || serviceBinding.Status.Conditions[0].Status == v1beta1.ConditionUnknown {
			return false
		}
		g.Expect(serviceBinding.Status.Conditions[0].Status).To(Equal(v1beta1.ConditionTrue))
		return true
	})

}

func waitForCompletion(g *GomegaWithT, test func() bool) {
	valid := false
	expiry := time.Now().Add(time.Duration(300) * time.Second)
	for !valid {
		valid = test()
		if !valid {
			fmt.Println("Not ready - waiting 10s...")
			time.Sleep(time.Duration(10) * time.Second)
			g.Expect(time.Now().Before(expiry)).To(BeTrue(), "Timeout expired")
		}
	}
}
