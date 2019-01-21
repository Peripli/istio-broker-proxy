package integration

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	. "github.com/onsi/gomega"
	"istio.io/istio/pilot/pkg/model"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
	"path"
	"strings"
	"testing"
	"text/template"
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

const service_instance_rabbitmq = `apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  name: rabbitmq-instance
spec:
  clusterServiceClassExternalName: rabbitmq
  clusterServicePlanExternalName: v3.7-dev`

const service_binding_rabbitmq = `apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  name: rabbitmq-binding
spec:
  instanceRef:
    name: rabbitmq-instance`

const client_config_postgres = `---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: client-postgres
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: client-postgres
    spec:
      volumes:
      - name: postgres-binding
        secret:
          secretName: postgres-binding
      containers:
      - name: client
        image: {{.HUB}}/client:latest
        command: ["/bin/sleep","infinity"]
        imagePullPolicy: Always
        volumeMounts:
        - mountPath: /etc/bindings/postgres
          name: postgres-binding`

const client_config_postgres_other_namespace = `---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: client-postgres
  namespace: integration-tests
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: client-postgres
    spec:
      volumes:
      - name: postgres-binding
        secret:
          secretName: postgres-binding
      containers:
      - name: client
        image: {{.HUB}}/client:latest
        command: ["/bin/sleep","infinity"]
        imagePullPolicy: Always
        volumeMounts:
        - mountPath: /etc/bindings/postgres
          name: postgres-binding`

const client_config_rabbitmq = `---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: client-rabbitmq
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: client-rabbitmq
    spec:
      volumes:
      - name: rabbitmq-binding
        secret:
          secretName: rabbitmq-binding
      containers:
      - name: client
        image: {{.HUB}}/client:latest
        command: ["/bin/sleep","infinity"]
        imagePullPolicy: Always
        volumeMounts:
        - mountPath: /etc/bindings/rabbitmq
          name: rabbitmq-binding`

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

type ClusterServiceClassObject struct {
	ApiVersion string                  `json:"apiVersion"`
	Kind       string                  `json:"kind"`
	Metadata   Metadata                `json:"metadata"`
	Spec       ClusterServiceClassSpec `json:"spec"`
}

type ClusterServiceClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []ClusterServiceClassObject `json:"items" protobuf:"bytes,2,rep,name=items"`
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

type ClusterServiceClassSpec struct {
	ExternalName             string `json:"externalName"`
	ClusterServiceBrokerName string `json:"clusterServiceBrokerName"`
}

var pgbenchOutput string
var pgbenchTime = 10

func init() {
	flag.StringVar(&pgbenchOutput, "pgbench-output", "", "Output file of postgres benchmark")
	flag.IntVar(&pgbenchTime, "pgbench-time", 10, "Duration of postgres benchmark test in seconds")
}

func TestPostgresServiceBinding(t *testing.T) {
	skipWithoutKubeconfigSet(t)

	g := NewGomegaWithT(t)
	kubectl := NewKubeCtl(g)

	createServiceBinding(kubectl, g, "postgres", service_instance, service_binding)

	podName := runClientPod(kubectl, client_config_postgres, "client-postgres")
	g.Expect(podName).To(ContainSubstring("client-postgres"))

	script := `  
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
		`
	basename := "test.sh"
	kubeCreateFile(kubectl, g, basename, script, podName)

	kubectl.Exec(podName, "-c", "client", "-i", "--", "bash", "test.sh")

}

func TestPostgresServiceBindingOtherNamespace(t *testing.T) {
	skipWithoutKubeconfigSet(t)

	g := NewGomegaWithT(t)
	kubectl := NewKubeCtl(g)

	kubectl.Delete("namespace", "integration-tests")
	kubectl.CreateNamespace("integration-tests")
	defer kubectl.Delete("namespace", "integration-tests")

	createServiceBinding(kubectl, g, "postgres", service_instance, service_binding)

	podName := runClientPod(kubectl, client_config_postgres_other_namespace, "client-postgres")
	g.Expect(podName).To(ContainSubstring("client-postgres"))

	script := `  
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
		`
	basename := "test.sh"
	kubeCreateFile(kubectl, g, basename, script, podName)

	kubectl.Exec(podName, "-c", "client", "-i", "--", "bash", "test.sh")

}

func TestPostgresBenchmark(t *testing.T) {
	skipWithoutKubeconfigSet(t)

	g := NewGomegaWithT(t)
	kubectl := NewKubeCtl(g)

	createServiceBinding(kubectl, g, "postgres", service_instance, service_binding)

	podName := runClientPod(kubectl, client_config_postgres, "client-postgres")
	g.Expect(podName).To(ContainSubstring("client-postgres"))

	script := fmt.Sprintf(`  
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
		    pgbench  -h $HOSTNAME  -p $PORT -U $USER -i -s 10 $DBNAME > /dev/null 2>&1
            pgbench  -h $HOSTNAME  -p $PORT -c 10 -T %d -j 4 -U $USER  $DBNAME | tee /tmp/pgbench.log
            exit 0
		  else
		    echo $OUTPUT
		    exit 1
		  fi
		done
		`, pgbenchTime)
	basename := "benchmark.sh"
	kubeCreateFile(kubectl, g, basename, script, podName)
	kubectl.Exec(podName, "-c", "client", "-i", "--", "bash", "benchmark.sh")

	if pgbenchOutput != "" {
		kubectl.run("cp", "default/"+podName+":/tmp/pgbench.log", pgbenchOutput)
	}
}

func runClientPod(kubectl *kubectl, config string, appName string) string {
	config = replaceHub(kubectl.g, config)
	clientConfigBody := []byte(config)
	kubectl.Apply(clientConfigBody)
	podName := kubectl.GetPod("-l", "app="+appName, "--field-selector=status.phase=Running")
	return podName
}

func replaceHub(g *GomegaWithT, config string) string {
	tmpl, err := template.New("replace-hub").Parse(config)
	g.Expect(err).NotTo(HaveOccurred())
	//A docker HUB is required to run these tests
	hub := os.Getenv("HUB")
	g.Expect(hub).NotTo(BeEmpty())
	writer := &bytes.Buffer{}
	tmpl.Execute(writer, map[string]string{"HUB": hub})
	return string(writer.Bytes())
}

func kubeCreateFile(kubectl *kubectl, g *GomegaWithT, basename string, script string, podName string) {
	fileName := path.Join(os.TempDir(), basename)
	file, err := os.Create(fileName)
	g.Expect(err).NotTo(HaveOccurred())
	file.Write([]byte(script))
	file.Close()
	kubectl.run("cp", fileName, "default/"+podName+":"+basename)
}

func TestServiceBindingIstioObjectsDeletedProperly(t *testing.T) {
	skipWithoutKubeconfigSet(t)

	g := NewGomegaWithT(t)
	kubectl := NewKubeCtl(g)

	bindId := createServiceBinding(kubectl, g, "postgres", service_instance, service_binding)

	kubectl.Delete("ServiceBinding", "postgres-binding")
	kubectl.Delete("ServiceInstance", "postgres-instance")

	var serviceEntries ServiceEntryList
	kubectl.List(&serviceEntries, "-n", "catalog")
	matchingIstioObjectCount := 0
	for _, serviceEntry := range serviceEntries.Items {
		if strings.Contains(serviceEntry.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(0))

	var virtualServices VirtualServiceList
	kubectl.List(&virtualServices, "-n", "catalog")

	for _, virtualService := range virtualServices.Items {

		if strings.Contains(virtualService.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(0))

	var gateways GatewayList
	kubectl.List(&gateways, "-n", "catalog")
	matchingIstioObjectCount = 0

	for _, gateway := range gateways.Items {

		if strings.Contains(gateway.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(0))

	var destinationRules DestinationruleList
	kubectl.List(&destinationRules, "-n", "catalog")
	matchingIstioObjectCount = 0

	for _, destinationRule := range destinationRules.Items {

		if strings.Contains(destinationRule.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(0))

}

func TestRabbitMqServiceBinding(t *testing.T) {
	skipWithoutKubeconfigSet(t)

	g := NewGomegaWithT(t)
	kubectl := NewKubeCtl(g)

	createServiceBinding(kubectl, g, "rabbitmq", service_instance_rabbitmq, service_binding_rabbitmq)

	podName := runClientPod(kubectl, client_config_rabbitmq, "client-rabbitmq")
	g.Expect(podName).To(ContainSubstring("client-rabbitmq"))

	script := `  
import pika
import time
while True:
  try:
    with open('/etc/bindings/rabbitmq/uri', 'r') as content_file:
      uri = content_file.read()
    print "Connecting to " + uri
    connection = pika.BlockingConnection(parameters=pika.URLParameters(uri))
    connection.close()
    break
  except (pika.exceptions.IncompatibleProtocolError, IOError, pika.exceptions.ConnectionClosed), e:
    print "Try again"
    time.sleep(10)
print "Connection to rqabbitmq was successful"
`
	basename := "test.py"
	kubeCreateFile(kubectl, g, basename, script, podName)

	kubectl.Exec(podName, "-c", "client", "-i", "--", "/usr/bin/python2.7", basename)
}

func createServiceBinding(kubectl *kubectl, g *GomegaWithT, name string, serviceConfig string, bindingConfig string) string {
	// Test if list of available services is not empty
	var classes v1beta1.ClusterServiceClassList
	kubectl.List(&classes)
	g.Expect(classes.Items).NotTo(BeEmpty(), "List of available services in OSB should not be empty")
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
	var services v1.ServiceList
	kubectl.List(&services, "--all-namespaces=true")
	g.Expect(services.Items).NotTo(BeEmpty(), "List of available services in OSB should not be empty")
	g.Expect(serviceExists(services, bindId)).To(BeTrue())
	var serviceEntries ServiceEntryList
	kubectl.List(&serviceEntries, "--all-namespaces=true")
	matchingServiceEntryExists := false
	for _, serviceEntry := range serviceEntries.Items {
		if strings.Contains(serviceEntry.Metadata.Name, bindId) {
			matchingServiceEntryExists = true
		}
	}
	g.Expect(matchingServiceEntryExists).To(BeTrue())
	var virtualServices VirtualServiceList
	kubectl.List(&virtualServices, "--all-namespaces=true")
	matchingIstioObjectCount := 0
	for _, virtualService := range virtualServices.Items {

		if strings.Contains(virtualService.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(2))
	var gateways GatewayList
	kubectl.List(&gateways, "--all-namespaces=true")
	matchingIstioObjectCount = 0
	for _, gateway := range gateways.Items {

		if strings.Contains(gateway.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(1))
	var destinationRules DestinationruleList
	kubectl.List(&destinationRules, "--all-namespaces=true")
	matchingIstioObjectCount = 0
	for _, destinationRule := range destinationRules.Items {

		if strings.Contains(destinationRule.Metadata.Name, bindId) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(2))
	return bindId
}

func serviceExists(services v1.ServiceList, bindId string) bool {
	for _, service := range services.Items {
		if strings.Contains(service.Name, bindId) {
			return true
		}
	}
	return false
}

func waitForCompletion(g *GomegaWithT, test func() bool) {
	valid := false
	expiry := time.Now().Add(time.Duration(20) * time.Minute)
	for !valid {
		valid = test()
		if !valid {
			log.Println("Not ready - waiting 10s...")
			time.Sleep(time.Duration(10) * time.Second)
			g.Expect(time.Now().Before(expiry)).To(BeTrue(), "Timeout expired")
		}
	}
}
