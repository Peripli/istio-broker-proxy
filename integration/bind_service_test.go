package integration

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	. "github.com/onsi/gomega"
	"io/ioutil"
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
  name: {{ .name }}
spec:
  clusterServiceClassExternalName: postgresql
  clusterServicePlanExternalName: v9.4-dev`

const service_binding = `apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  name: {{ .name }}
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
      imagePullSecrets:
      - name: image-pull-secret
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
      imagePullSecrets:
      - name: image-pull-secret
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
      imagePullSecrets:
      - name: image-pull-secret
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

func skipWithoutEnableBackingServicesSet(t *testing.T) {
	if os.Getenv("ENABLE_BACKING_SERVICES") == "" {
		t.Skip("ENABLE_BACKING_SERVICES not set, skipping integration test.")
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
	skipWithoutEnableBackingServicesSet(t)

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
	skipWithoutEnableBackingServicesSet(t)

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
	skipWithoutEnableBackingServicesSet(t)

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
	kubectl.RolloutStatus(appName)
	podName := kubectl.GetPod("-l", "app="+appName, "--field-selector=status.phase=Running")
	return podName
}

func replaceHub(g *GomegaWithT, config string) string {
	tmpl, err := template.New("replace-hub").Parse(config)
	g.Expect(err).NotTo(HaveOccurred())
	//A docker HUB is required to run these tests
	hub := os.Getenv("HUB")
	fmt.Printf("Using HUB=%s", hub)
	g.Expect(hub).NotTo(BeEmpty())
	writer := &bytes.Buffer{}
	tmpl.Execute(writer, map[string]string{"HUB": hub})
	return string(writer.Bytes())
}

func replaceName(g *GomegaWithT, config string, name string) string {
	tmpl, err := template.New("replace-hub").Parse(config)
	g.Expect(err).NotTo(HaveOccurred())
	//A docker HUB is required to run these tests
	writer := &bytes.Buffer{}
	tmpl.Execute(writer, map[string]string{"name": name})
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

func kubeFetchFile(kubectl *kubectl, g *GomegaWithT, basename string, podName string, containerName string) string {
	fileName := path.Join(os.TempDir(), basename)
	out := kubectl.run("cp", "-c", containerName, "default/"+podName+":"+basename, fileName)
	fmt.Print(string(out))
	file, err := ioutil.ReadFile(fileName)
	g.Expect(err).NotTo(HaveOccurred())
	return string(file)
}

func TestRabbitMqServiceBinding(t *testing.T) {
	skipWithoutKubeconfigSet(t)
	skipWithoutEnableBackingServicesSet(t)

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
print "Connection to rabbitmq was successful"
`
	basename := "test.py"
	kubeCreateFile(kubectl, g, basename, script, podName)

	kubectl.Exec(podName, "-c", "client", "-i", "--", "/usr/bin/python2.7", basename)
}

func createServiceBinding(kubectl *kubectl, g *GomegaWithT, name string, serviceConfig string, bindingConfig string) string {
	// Test if list of available services is not empty
	var classes v1beta1.ClusterServiceClassList
	kubectl.List(&classes)
	instanceName := name + "-instance"
	bindingName := name + "-binding"
	g.Expect(classes.Items).NotTo(BeEmpty(), "List of available services in OSB should not be empty")
	kubectl.Delete("ServiceBinding", bindingName)
	kubectl.Delete("ServiceInstance", instanceName)
	kubectl.Apply([]byte(replaceName(g, serviceConfig, instanceName)))
	waitForServiceInstance(kubectl, g, name)
	kubectl.Apply([]byte(replaceName(g, bindingConfig, bindingName)))
	serviceBinding := waitForServiceBinding(kubectl, g, name)
	bindID := serviceBinding.Spec.ExternalID
	var services v1.ServiceList
	kubectl.List(&services, "--all-namespaces=true")
	g.Expect(services.Items).NotTo(BeEmpty(), "List of available services in OSB should not be empty")
	g.Expect(serviceExists(services, bindID)).To(BeTrue())
	var serviceEntries ServiceEntryList
	kubectl.List(&serviceEntries, "--all-namespaces=true")
	matchingServiceEntryExists := false
	for _, serviceEntry := range serviceEntries.Items {
		if strings.Contains(serviceEntry.Metadata.Name, bindID) {
			matchingServiceEntryExists = true
		}
	}
	g.Expect(matchingServiceEntryExists).To(BeTrue())
	var virtualServices VirtualServiceList
	kubectl.List(&virtualServices, "--all-namespaces=true")
	matchingIstioObjectCount := 0
	for _, virtualService := range virtualServices.Items {

		if strings.Contains(virtualService.Metadata.Name, bindID) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(2))
	var gateways GatewayList
	kubectl.List(&gateways, "--all-namespaces=true")
	matchingIstioObjectCount = 0
	for _, gateway := range gateways.Items {

		if strings.Contains(gateway.Metadata.Name, bindID) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(1))
	var destinationRules DestinationruleList
	kubectl.List(&destinationRules, "--all-namespaces=true")
	matchingIstioObjectCount = 0
	for _, destinationRule := range destinationRules.Items {

		if strings.Contains(destinationRule.Metadata.Name, bindID) {
			matchingIstioObjectCount += 1
		}
	}
	g.Expect(matchingIstioObjectCount).To(Equal(2))
	return bindID
}

func serviceExists(services v1.ServiceList, bindID string) bool {
	for _, service := range services.Items {
		if strings.Contains(service.Name, bindID) {
			return true
		}
	}
	return false
}

func waitForServiceBinding(kubectl *kubectl, g *GomegaWithT, namePrefix string) v1beta1.ServiceBinding {
	var serviceBinding v1beta1.ServiceBinding

	waitForCompletion(g, func() (bool, string) {
		name := namePrefix + "-binding"
		kubectl.Read(&serviceBinding, name)
		statusLen := len(serviceBinding.Status.Conditions)
		if statusLen == 0 {
			return false, ""
		}
		condition := serviceBinding.Status.Conditions[statusLen-1]
		reason := string(condition.Reason)
		if condition.Status != v1beta1.ConditionTrue {
			return false, reason
		}

		if v1beta1.ServiceBindingConditionReady != condition.Type {
			log.Println(kubectl.Describe(&serviceBinding, name))
			g.Expect(condition.Type).To(Equal(v1beta1.ServiceBindingConditionReady), fmt.Sprintf("reason: %s", string(condition.Reason)))
		}

		return true, reason
	}, "servicebinding")

	return serviceBinding
}

func waitForServiceInstance(kubectl *kubectl, g *GomegaWithT, namePrefix string) v1beta1.ServiceInstance {
	var serviceInstance v1beta1.ServiceInstance

	waitForCompletion(g, func() (bool, string) {
		name := namePrefix+"-instance"
		kubectl.Read(&serviceInstance, name)
		statusLen := len(serviceInstance.Status.Conditions)
		if statusLen == 0 {
			return false, ""
		}

		condition := serviceInstance.Status.Conditions[statusLen-1]
		reason := string(condition.Reason)
		log.Printf("reason: %s", reason)
		if condition.Status != v1beta1.ConditionTrue {
			return false, reason
		}

		if v1beta1.ServiceInstanceConditionReady != condition.Type {
			log.Println(kubectl.Describe(&serviceInstance, name))
			g.Expect(condition.Type).To(Equal(v1beta1.ServiceInstanceConditionReady))
		}
		return true, reason
	}, "serviceinstance")
	return serviceInstance
}

func waitForCompletion(g *GomegaWithT, test func() (bool, string), name string) {
	valid := false
	expiry := time.Now().Add(MAX_WAITING_TIME)
	for !valid {
		var lastReason string
		valid, lastReason = test()
		if !valid {
			log.Printf("Not ready yet - waiting...")
			if lastReason != "" {
				log.Printf("Reason: %s", lastReason)
			}
			g.Expect(lastReason).NotTo(ContainSubstring("Failed"))
			g.Expect(lastReason).NotTo(ContainSubstring("Error"))
			time.Sleep(ITERATION_WAITING_TIME)
			g.Expect(time.Now().Before(expiry)).To(BeTrue(), fmt.Sprintf("Timeout expired while waiting for: %s.\n Reason: %s", name, lastReason))
		}
	}
}
