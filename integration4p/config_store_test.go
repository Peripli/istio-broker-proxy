package integration4p

import (
	"os"
	"testing"

	"github.com/Peripli/istio-broker-proxy/pkg/router"
	. "github.com/onsi/gomega"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/Peripli/istio-broker-proxy/integration"
)

func skipWithoutKubeconfigSet(t *testing.T) {
	if os.Getenv("KUBECONFIG") == "" {
		t.Skip("KUBECONFIG not set, skipping integration test.")
	}
}

func TestKubernetesCreateService(t *testing.T) {
	skipWithoutKubeconfigSet(t)

	g := NewGomegaWithT(t)
	kubectl := integration.NewKubeCtl(g)
	service := &v1.Service{Spec: v1.ServiceSpec{Ports: []v1.ServicePort{{Port: 5555, TargetPort: intstr.FromInt(5555)}}}}
	service.Name = "test-config-store"
	kubectl.Delete("Service", service.Name)

	configStore := router.NewExternKubeConfigStore("default")
	service, err := configStore.CreateService(service)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(service.Spec.ClusterIP).ToNot(BeEmpty())

}
