package binding

import (
	"fmt"
	. "github.com/onsi/gomega"
	"istio.io/api/networking/v1alpha3"
	"testing"
)

const example_string = `kind: Endpoints
apiVersion: v1
metadata:
name: my-service
subsets:
	- addresses: "postgres.services.cf.dev01.aws.istio.sapcloud.io"
	- ip: 1.2.3.4
ports:
	- port: 9000`

//func TestCreateResponseWithIstioData(t *testing.T) {
//	g := NewGomegaWithT(t)
//	var request http.Request
//	body := request.Body
//	response, error := addIstioDataToResponse(request, body)
//	fmt.Printf("%v", response)
//
//	emptyResponse := http.Response{}
//	g.Expect(error).NotTo(HaveOccurred())
//	g.Expect(response).NotTo(Equal(emptyResponse))
//}

func TestNetworkDataExists(t *testing.T) {
	g := NewGomegaWithT(t)

	endpointsBody := []byte(example_string)
	networkData, err := joinNetworkData(endpointsBody)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(networkData).NotTo(BeNil())
	g.Expect(networkData.profileId).NotTo(Equal(""))
	g.Expect(networkData.data).NotTo(BeNil())
	g.Expect(networkData.data.providerId).NotTo(Equal(""))
	g.Expect(networkData.data.endpoints["host"]).NotTo(Equal(""))
	g.Expect(networkData.data.endpoints["port"]).NotTo(Equal(""))

}

func TestNetworkDataContainsData(t *testing.T) {
	g := NewGomegaWithT(t)

	endpointsBody := []byte(example_string)
	networkData, err := joinNetworkData(endpointsBody)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(networkData).NotTo(BeNil())
	g.Expect(networkData.profileId).NotTo(Equal(""))
	g.Expect(networkData.data).NotTo(BeNil())
	g.Expect(networkData.data.providerId).NotTo(Equal(""))
}

func TestExtractDataFromEndpoints(t *testing.T) {
	g := NewGomegaWithT(t)

	//example_string := `"network_data": {
	//"network_profile_id": "urn:istio:public",
	//"data": {
	//  "provider_id": "spiffe://ingress.services.cf.dev01.aws.istio.sapcloud.io",
	//  "endpoints": [{
	//    "host": "postgres.services.cf.dev01.aws.istio.sapcloud.io",
	//    "port": 9000
	//  }]
	//}`

	endpointsBody := []byte(example_string)
	host, port := extractDataFromEndpoints(endpointsBody)

	endpointAddress := "postgres.services.cf.dev01.aws.istio.sapcloud.io"
	endpointPorts := make(map[string]uint32)
	portName := "port"
	endpointPorts[portName] = 9000
	endpoint := v1alpha3.ServiceEntry_Endpoint{Address: endpointAddress, Ports: endpointPorts}

	g.Expect(host).To(Equal(endpoint.Address))
	portString := fmt.Sprint(endpoint.Ports["port"])
	g.Expect(port).To(Equal(portString))
}
