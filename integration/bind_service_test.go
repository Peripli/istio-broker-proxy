package integration

import (
	"fmt"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	. "github.com/onsi/gomega"
	"istio.io/istio/pilot/pkg/model"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path"
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

const client_config = `---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: client
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: client
    spec:
      volumes:
      - name: postgres-binding
        secret:
          secretName: postgres-binding
      containers:
      - name: client
        image: gcr.io/sap-se-gcp-istio-dev/client:latest
        command: ["/bin/sleep","infinity"]
        imagePullPolicy: Always
        volumeMounts:
        - mountPath: /etc/bindings/postgres
          name: postgres-binding`

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
	CreationTimestamp string `json:"creationTimestamp"`
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
	kubectl.Apply([]byte(service_binding))
	var serviceBinding v1beta1.ServiceBinding
	waitForCompletion(g, func() bool {
		kubectl.Read(&serviceBinding, "postgres-binding")
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
		}
	}
	g.Expect(matchingServiceEntryExists).To(BeTrue())

	var virtualServices VirtualServiceList
	kubectl.List(&virtualServices, "-n", "catalog")
	matchingIstioObjectCount := 0

	for _, virtualService := range virtualServices.Items {

		if strings.Contains(virtualService.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(2))

	var gateways GatewayList
	kubectl.List(&gateways, "-n", "catalog")
	matchingIstioObjectCount = 0

	for _, gateway := range gateways.Items {

		if strings.Contains(gateway.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(1))

	var destinationRules DestinationruleList
	kubectl.List(&destinationRules, "-n", "catalog")
	matchingIstioObjectCount = 0

	for _, destinationRule := range destinationRules.Items {

		if strings.Contains(destinationRule.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(2))

	clientConfigBody := []byte(client_config)
	kubectl.Apply(clientConfigBody)

	podName := kubectl.GetPod("-l", "app=client", "--field-selector=status.phase=Running")
	g.Expect(podName).To(ContainSubstring("client"))

	fileName := path.Join(os.TempDir(), "test.sh")
	file, err := os.Create(fileName)
	g.Expect(err).NotTo(HaveOccurred())
	file.Write([]byte(`  
while true; do
  export PGPASSWORD=$(cat /etc/bindings/postgres/password)
  HOSTNAME=$(cat /etc/bindings/postgres/hostname)
  PORT=$(cat /etc/bindings/postgres/port)
  DBNAME=$(cat /etc/bindings/postgres/dbname)
  USER=$(cat /etc/bindings/postgres/username)
  OUTPUT=$(psql -h $HOSTNAME  -p $PORT -c 'SELECT 1'  $DBNAME $USER 2>&1)
  echo $OUTPUT >> /tmp/psql.txt
  if [[ $OUTPUT == *"server closed the connection unexpectedly"* ]]; then
    echo "Try again!"
    sleep 10
  elif [[ $OUTPUT == *"(1 row)"* ]]; then
    break
  else
    echo $OUTPUT
    exit 1
  fi
done
`))
	file.Close()
	kubectl.run("cp", fileName, "default/"+podName+":test.sh")

	kubectl.Exec(podName, "-c", "client", "-ti", "--", "bash", "test.sh")

}

func TestServiceBindingIstioObjectsDeletedProperly(t *testing.T) {
	skipWithoutKubeconfigSet(t)

	g := NewGomegaWithT(t)
	kubectl := NewKubeCtl(g)

	// Test if list of available services is not empty
	var classes v1beta1.ClusterServiceClassList
	kubectl.List(&classes)
	g.Expect(classes.Items).NotTo(BeEmpty(), "List of available services in OSB should not be empty")

	kubectl.Apply([]byte(service_instance))
	var serviceInstance v1beta1.ServiceInstance
	waitForCompletion(g, func() bool {
		kubectl.Read(&serviceInstance, "postgres-instance")
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
	kubectl.Apply([]byte(service_binding))
	var serviceBinding v1beta1.ServiceBinding
	waitForCompletion(g, func() bool {
		kubectl.Read(&serviceBinding, "postgres-binding")
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
	matchingIstioObjectCount := 0
	for _, serviceEntry := range serviceEntries.Items {
		if strings.Contains(serviceEntry.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(BeNumerically(">=", 1))

	var virtualServices VirtualServiceList
	kubectl.List(&virtualServices, "-n", "catalog")

	for _, virtualService := range virtualServices.Items {

		if strings.Contains(virtualService.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(BeNumerically(">=", 1))

	var gateways GatewayList
	kubectl.List(&gateways, "-n", "catalog")
	matchingIstioObjectCount = 0

	for _, gateway := range gateways.Items {

		if strings.Contains(gateway.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(BeNumerically(">=", 1))

	var destinationRules DestinationruleList
	kubectl.List(&destinationRules, "-n", "catalog")
	matchingIstioObjectCount = 0

	for _, destinationRule := range destinationRules.Items {

		if strings.Contains(destinationRule.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(BeNumerically(">=", 1))

	kubectl.Delete("ServiceBinding", "postgres-binding")
	kubectl.Delete("ServiceInstance", "postgres-instance")

	kubectl.List(&serviceEntries, "-n", "catalog")
	matchingIstioObjectCount = 0
	for _, serviceEntry := range serviceEntries.Items {
		if strings.Contains(serviceEntry.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(0))

	kubectl.List(&virtualServices, "-n", "catalog")

	for _, virtualService := range virtualServices.Items {

		if strings.Contains(virtualService.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(0))

	kubectl.List(&gateways, "-n", "catalog")
	matchingIstioObjectCount = 0

	for _, gateway := range gateways.Items {

		if strings.Contains(gateway.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(0))

	kubectl.List(&destinationRules, "-n", "catalog")
	matchingIstioObjectCount = 0

	for _, destinationRule := range destinationRules.Items {

		if strings.Contains(destinationRule.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(0))

}

func waitForCompletion(g *GomegaWithT, test func() bool) {
	valid := false
	expiry := time.Now().Add(time.Duration(20) * time.Minute)
	for !valid {
		valid = test()
		if !valid {
			fmt.Println("Not ready - waiting 10s...")
			time.Sleep(time.Duration(10) * time.Second)
			g.Expect(time.Now().Before(expiry)).To(BeTrue(), "Timeout expired")
		}
	}
}
