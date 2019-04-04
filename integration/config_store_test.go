package integration

import (
	"github.com/Peripli/istio-broker-proxy/pkg/config"
	"istio.io/istio/pilot/pkg/config/kube/crd"
	"k8s.io/client-go/tools/clientcmd"
	"testing"

	"github.com/Peripli/istio-broker-proxy/pkg/router"
	. "github.com/onsi/gomega"
	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/model"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestKubernetesCreateService(t *testing.T) {
	skipWithoutKubeconfigSet(t)

	g := NewGomegaWithT(t)
	kubectl := NewKubeCtl(g)
	service := &v1.Service{Spec: v1.ServiceSpec{Ports: []v1.ServicePort{{Port: 5555, TargetPort: intstr.FromInt(5555)}}}}
	service.Name = "test-config-store"
	kubectl.Delete("Service", service.Name)

	configStore := router.NewExternKubeConfigStore("default")
	service, err := configStore.CreateService(service)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(service.Spec.ClusterIP).ToNot(BeEmpty())

}

func TestKubernetesCreateIstioConfig(t *testing.T) {
	skipWithoutKubeconfigSet(t)

	g := NewGomegaWithT(t)
	kubectl := NewKubeCtl(g)

	cfg := model.Config{Spec: &v1alpha3.ServiceEntry{
		Hosts: []string{"host.cluster.local"},
		Ports: []*v1alpha3.Port{{
			Number:   6666,
			Name:     "test-port",
			Protocol: "TLS"}},
		Resolution: v1alpha3.ServiceEntry_DNS},
		ConfigMeta: model.ConfigMeta{Labels: map[string]string{"foo": "bar"}},
	}
	cfg.Type = "service-entry"
	cfg.Name = "test-service-entry"
	cfg.Group = "networking.istio.io"
	cfg.Version = "v1alpha3"
	kubectl.Delete("ServiceEntry", cfg.Name)
	g.Expect(checkIfServiceExists(kubectl, "foo=bar")).To(BeFalse())

	clientcmd.ClusterDefaults.Server = ""
	configStore := router.NewExternKubeConfigStore("default")
	err := configStore.CreateIstioConfig("", cfg)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(checkIfServiceExists(kubectl, "foo=bar")).To(BeTrue())

	kubectl.Delete("ServiceEntry", cfg.Name)
	g.Expect(checkIfServiceExists(kubectl, "foo=bar")).To(BeFalse())
}

func TestKubernetesCreateIstioObjects(t *testing.T) {
	skipWithoutKubeconfigSet(t)

	g := NewGomegaWithT(t)
	kubectl := NewKubeCtl(g)
	clientcmd.ClusterDefaults.Server = ""
	configStore := router.NewExternKubeConfigStore("not-used")

	configurations := config.CreateEntriesForExternalServiceClient("svc-0-f1b32107-c8a5-11e8-b8be-02caceffa7f1", "test.istio.my.arbitrary.domain.io", "1.1.1.1", 1234, "catalog", "my.arbitrary.domain.io")

	g.Expect(configurations).To(HaveLen(6))

	deleteClientObjects(kubectl, configurations)

	for _, configuration := range configurations {

		err := configStore.CreateIstioConfig("", configuration)
		g.Expect(err).NotTo(HaveOccurred(), "error creating %#v\n", configuration)
	}

	deleteClientObjects(kubectl, configurations)
}

func deleteClientObjects(kubectl *kubectl, configurations []model.Config) {
	kubectl.DeleteWithNamespace("ServiceEntry", configurations[0].Name, "catalog")
	kubectl.DeleteWithNamespace("VirtualService", configurations[1].Name, "catalog")
	kubectl.DeleteWithNamespace("VirtualService", configurations[2].Name, "catalog")
	kubectl.DeleteWithNamespace("Gateway", configurations[3].Name, "catalog")
	kubectl.DeleteWithNamespace("DestinationRule", configurations[4].Name, "catalog")
	kubectl.DeleteWithNamespace("DestinationRule", configurations[5].Name, "catalog")
}

func checkIfServiceExists(kubectl *kubectl, label string) bool {
	var serviceEntries crd.ServiceEntryList
	kubectl.List(&serviceEntries, "-l", label)
	return len(serviceEntries.GetItems()) > 0
}
