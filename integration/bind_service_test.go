// +build integration

package integration

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	. "github.com/onsi/gomega"
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

func TestServiceBindingIsSuccessful(t *testing.T) {

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
		return len(serviceInstance.Status.Conditions) > 0 && serviceInstance.Status.Conditions[0].Status == v1beta1.ConditionTrue
	})
	kubectl.Apply([]byte(service_binding))
	var status v1beta1.ServiceBinding
	waitForCompletion(g, func() bool {
		kubectl.Read(&status, "postgres-binding")
		return len(status.Status.Conditions) > 0 && status.Status.Conditions[0].Status == v1beta1.ConditionTrue
	})

}

func waitForCompletion(g *GomegaWithT, test func() bool) {
	valid := false
	expiry := time.Now().Add(time.Duration(60) * time.Second)
	for !valid {
		valid = test()
		if !valid {
			time.Sleep(time.Duration(5) * time.Second)
			g.Expect(time.Now().Before(expiry)).To(BeTrue(), "Timeout expired")
		}
	}
}
