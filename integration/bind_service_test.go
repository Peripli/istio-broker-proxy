package integration

import (
	"fmt"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	. "github.com/onsi/gomega"
	"istio.io/istio/pilot/pkg/model"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"strings"
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

type IstioObjectList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// List of services
	Items []IstioObject `json:"items" protobuf:"bytes,2,rep,name=items"`
}

type IstioObject struct {
	ApiVersion string   `json:"apiVersion"`
	Kind       string   `json:"kind"`
	Metadata   Metadata `json:"metadata"`
	Spec       Spec     `json:"spec"`
}

// ServiceList holds a list of services.
type ServiceEntryList struct {
	IstioObjectList
}

type VirtualServiceList struct {
	IstioObjectList
}

type DestinationruleList struct {
	IstioObjectList
}

type GatewayList struct {
	IstioObjectList
}

type Metadata struct {
	CreationTimestamp string `json:"creationTimestamp""`
	Generation        int    `json:"generation"`
	Name              string `json:"name"`
	Namespace         string `json:"namespace"`
	ResourceVersion   string `json:"resourceVersion"`
	SelfLink          string `json:"selfLink"`
	UId               string `json:"uid"`
}

type Spec struct {
	Hosts      []string     `json:"hosts"`
	Ports      []model.Port `json:"ports"`
	Resoultion string       `json:"resolution"`
}

func TestServiceBindingIstioObjectsCreated(t *testing.T) {
	skipWithoutKubeconfigSet(t)

	g := NewGomegaWithT(t)
	kubectl := NewKubeCtl(g)

	// Test if list of available services is not empty
	var classes v1beta1.ClusterServiceClassList
	kubectl.List(&classes)
	g.Expect(classes.Items).NotTo(BeEmpty(), "List of available services in OSB should not be empty")

	kubectl.Delete("ServiceBinding", "postgres-binding")
	kubectl.Delete("ServiceInstance", "postgres-instance")

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
	bindId := serviceBinding.Spec.ExternalID
	var services v1.ServiceList
	kubectl.List(&services, "-n", "catalog")
	g.Expect(services.Items).NotTo(BeEmpty(), "List of available services in OSB should not be empty")
	matchingServiceInstanceExists := false
	for _, service := range services.Items {
		if strings.Contains(service.Name, bindId) {
			matchingServiceInstanceExists = true
		}
	}
	g.Expect(matchingServiceInstanceExists).To(BeTrue())

	matchingServiceInstanceExists = false
	for _, service := range services.Items {
		if strings.Contains(service.Name, "noPropperBindID") {
			matchingServiceInstanceExists = true
		}
	}
	g.Expect(matchingServiceInstanceExists).To(BeFalse())

	var serviceEntries ServiceEntryList
	kubectl.List(&serviceEntries, "-n", "catalog")
	matchingServiceEntryExists := false
	for _, serviceEntry := range serviceEntries.Items {
		if strings.Contains(serviceEntry.Metadata.Name, bindId) {
			matchingServiceEntryExists = true
			kubectl.Delete("ServiceEntry", serviceEntry.Metadata.Name)
		}
	}
	g.Expect(matchingServiceEntryExists).To(BeTrue())

	var virtualServices VirtualServiceList
	kubectl.List(&virtualServices, "-n", "catalog")
	matchingIstioObjectCount := 0

	for _, virtualService := range virtualServices.Items {

		if strings.Contains(virtualService.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
			kubectl.Delete("VirtualService", virtualService.Metadata.Name)
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(2))

	var gateways GatewayList
	kubectl.List(&gateways, "-n", "catalog")
	matchingIstioObjectCount = 0

	for _, gateway := range gateways.Items {

		if strings.Contains(gateway.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
			kubectl.Delete("Gateway", gateway.Metadata.Name)
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(1))

	var destinationRules DestinationruleList
	kubectl.List(&destinationRules, "-n", "catalog")
	matchingIstioObjectCount = 0

	for _, destinationRule := range destinationRules.Items {

		if strings.Contains(destinationRule.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
			kubectl.Delete("DestinationRule", destinationRule.Metadata.Name)
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(2))

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
