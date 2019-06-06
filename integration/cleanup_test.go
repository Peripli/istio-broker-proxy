package integration

import (
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/onsi/gomega"
	"k8s.io/api/core/v1"
	"log"
	"os"
	"regexp"
	"testing"
)

type counter struct {
	skipped int
	deleted int
}

func TestCleanupOrphanedObjects(t *testing.T) {
	skipWithoutKubeconfigSet(t)
	if os.Getenv("CLEANUP_ORPHANED_OBJECTS") != "true" {
		t.Skip("Skipping cleanup of orphaned routes.")
	}
	var g = gomega.NewGomegaWithT(t)
	var kubectl = NewKubeCtl(g)

	var serviceBindings v1beta1.ServiceBindingList
	var virtualServiceList VirtualServiceList
	var serviceEntryList ServiceEntryList
	var destinationruleList DestinationruleList
	var gatewayList GatewayList
	var serviceList v1.ServiceList

	counter := counter{0, 0}

	kubectl.List(&serviceBindings)
	kubectl.List(&virtualServiceList, "--all-namespaces")
	kubectl.List(&serviceEntryList, "--all-namespaces")
	kubectl.List(&destinationruleList, "--all-namespaces")
	kubectl.List(&gatewayList, "--all-namespaces")
	kubectl.List(&serviceList, "--all-namespaces")

	cleanupIstioObjects(&virtualServiceList.IstioObjectList, serviceBindings, kubectl, &counter)
	cleanupIstioObjects(&serviceEntryList.IstioObjectList, serviceBindings, kubectl, &counter)
	cleanupIstioObjects(&destinationruleList.IstioObjectList, serviceBindings, kubectl, &counter)
	cleanupIstioObjects(&gatewayList.IstioObjectList, serviceBindings, kubectl, &counter)
	cleanupServices(&serviceList, serviceBindings, kubectl, &counter)

	log.Printf("Deleted %d orphaned objects, keeping %d generated objects (still in use)", counter.deleted, counter.skipped)
}

func cleanupIstioObjects(istioObjectList *IstioObjectList, serviceBindings v1beta1.ServiceBindingList, kubectl *kubectl, counter *counter) {
	var generatedObjectRegex = regexp.MustCompile(
		"^(istio-egressgateway-|egressgateway-|sidecar-to-egress-|mesh-to-egress-|egress-gateway-)?svc-" + // prefix
			"[0-9]+-([0-9a-f]{8}(-[0-9a-f]{4}){4}[0-9a-f]{8})" + // index '-' uuid
			"(-service)?$") // suffix

	for _, istioObject := range istioObjectList.Items {

		groups := generatedObjectRegex.FindStringSubmatch(istioObject.Metadata.Name)
		if groups != nil {
			uuid := groups[2]

			found := false
			for _, serviceBinding := range serviceBindings.Items {
				found = found || (serviceBinding.Spec.ExternalID == uuid)
			}

			if !found {
				log.Printf("Delete %s: %s\n", istioObject.Kind, istioObject.Metadata.Name)
				kubectl.DeleteWithNamespace(istioObject.Kind, istioObject.Metadata.Name, istioObject.Metadata.Namespace)
				counter.deleted++
			} else {
				log.Printf("Keep %s: %s\n", istioObject.Kind, istioObject.Metadata.Name)
				counter.skipped++
			}
		}
	}
}

func cleanupServices(serviceList *v1.ServiceList, serviceBindings v1beta1.ServiceBindingList, kubectl *kubectl, counter *counter) {
	var generatedServiceRegex = regexp.MustCompile(
		"^svc-" + // prefix
			"[0-9]+-([0-9a-f]{8}(-[0-9a-f]{4}){4}[0-9a-f]{8})$") // index '-' uuid

	for _, service := range serviceList.Items {

		groups := generatedServiceRegex.FindStringSubmatch(service.Name)
		if groups != nil {
			uuid := groups[1]

			found := false
			for _, serviceBinding := range serviceBindings.Items {
				found = found || (serviceBinding.Spec.ExternalID == uuid)
			}

			if !found {
				log.Printf("Delete %s: %s", service.Kind, service.Name)
				kubectl.DeleteWithNamespace(service.Kind, service.Name, service.Namespace)
				counter.deleted++
			} else {
				log.Printf("Keep %s: %s", service.Kind, service.Name)
				counter.skipped++
			}
		}
	}
}
