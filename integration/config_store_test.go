package integration

import (
	"fmt"
	. "github.com/onsi/gomega"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/router"
	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/model"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

func TestKubernetesCreateService(t *testing.T) {
	skipWithoutKubeconfigSet(t)

	g := NewGomegaWithT(t)
	kubectl := NewKubeCtl(g)
	service := &v1.Service{Spec: v1.ServiceSpec{Ports: []v1.ServicePort{{Port: 5555, TargetPort: intstr.FromInt(5555)}}}}
	service.Name = "test-kubernetes"
	kubectl.Delete("Service", service.Name)

	kubernetes := router.NewExternKubeConfigStore("default")
	service, err := kubernetes.CreateService(service)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(service.Spec.ClusterIP).ToNot(BeEmpty())

}

func TestKubernetesCreateObject(t *testing.T) {
	t.Skip("Not working")
	skipWithoutKubeconfigSet(t)

	g := NewGomegaWithT(t)
	kubectl := NewKubeCtl(g)

	cfg := model.Config{Spec: &v1alpha3.ServiceEntry{
		Hosts: []string{"host.cluster.local"},
		Ports: []*v1alpha3.Port{{
			Number:   6666,
			Name:     "test-port",
			Protocol: "TLS"}},
		Resolution: v1alpha3.ServiceEntry_DNS}}
	cfg.Type = "service-entry"
	cfg.Name = "test-service-entry"

	fmt.Println(cfg.Group)
	fmt.Println(cfg.Version)

	kubectl.Delete("ServiceEntry", cfg.Name)

	kubernetes := router.NewExternKubeConfigStore("default")
	err := kubernetes.CreateIstioConfig(cfg)
	g.Expect(err).ToNot(HaveOccurred())

}
