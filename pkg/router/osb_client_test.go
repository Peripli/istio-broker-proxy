package router

import (
	"encoding/json"
	"github.com/Peripli/istio-broker-proxy/pkg/model"
	. "github.com/onsi/gomega"
	"net/http"
	"testing"
)

func TestAdaptCredentialsWithProxy(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := NewHandlerStubWithFunc(http.StatusOK, func(body []byte) []byte {
		var request model.AdaptCredentialsRequest
		err := json.Unmarshal(body, &request)
		g.Expect(err).NotTo(HaveOccurred())
		response, err := model.Adapt(request.Credentials, request.EndpointMappings)
		g.Expect(err).NotTo(HaveOccurred())
		responseBody, err := json.Marshal(response)
		g.Expect(err).NotTo(HaveOccurred())

		return responseBody
	})
	server, routerConfig := injectClientStub(handlerStub)
	defer server.Close()
	client := OsbClient{&RouterRestClient{routerConfig.HttpClientFactory(&http.Transport{}), make(map[string][]string), *routerConfig}}
	binding, err := client.AdaptCredentials("1234-4567", "7654-3210",
		model.PostgresCredentials{
			Port:     47637,
			Hostname: "10.11.241.0",
			Uri:      "postgres://mma4G8N0isoxe17v:redacted@10.11.241.0:47637/yLO2WoE0-mCcEppn",
		}.ToCredentials(),
		[]model.EndpointMapping{
			{
				Source: model.Endpoint{"10.11.241.0", 47637},
				Target: model.Endpoint{"postgres.catalog.svc.cluster.local", 5555},
			},
		})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(binding).NotTo(BeNil())
	credentials := binding.Credentials

	postgresCredentials, err := model.PostgresCredentialsFromCredentials(credentials)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(postgresCredentials.Port).To(Equal(5555))
	g.Expect(postgresCredentials.Hostname).To(Equal("postgres.catalog.svc.cluster.local"))
	g.Expect(postgresCredentials.Uri).To(Equal("postgres://mma4G8N0isoxe17v:redacted@postgres.catalog.svc.cluster.local:5555/yLO2WoE0-mCcEppn"))

}

func TestAdaptCredentialsWithBadRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := NewHandlerStub(http.StatusBadRequest, []byte(`{"error" : "myerror", "description" : "mydescription"}`))
	server, routerConfig := injectClientStub(handlerStub)
	defer server.Close()
	client := OsbClient{&RouterRestClient{routerConfig.HttpClientFactory(&http.Transport{}), make(map[string][]string), *routerConfig}}
	_, err := client.AdaptCredentials("1234-4567", "7654-3210",
		model.PostgresCredentials{}.ToCredentials(),
		[]model.EndpointMapping{})

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.(*model.HttpError).StatusCode).To(Equal(http.StatusBadRequest))
	g.Expect(err.(*model.HttpError).ErrorMsg).To(Equal("myerror"))
	g.Expect(err.(*model.HttpError).Description).To(Equal("mydescription"))
}

func TestAdaptCredentialsWithInvalidJson(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := NewHandlerStub(http.StatusOK, []byte(""))
	server, routerConfig := injectClientStub(handlerStub)
	defer server.Close()
	client := OsbClient{&RouterRestClient{routerConfig.HttpClientFactory(&http.Transport{}), make(map[string][]string), *routerConfig}}
	_, err := client.AdaptCredentials("1234-4567", "7654-3210",
		model.PostgresCredentials{}.ToCredentials(),
		[]model.EndpointMapping{})

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("Can't unmarshal response from"))
}

func TestGetCatalog(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := NewHandlerStubWithFunc(http.StatusOK, func(body []byte) []byte {
		return []byte(`{
  "services": [{
    "id": "id",
    "name": "name" }]
}`)
	})
	server, routerConfig := injectClientStub(handlerStub)
	defer server.Close()
	client := OsbClient{&RouterRestClient{routerConfig.HttpClientFactory(&http.Transport{}), make(map[string][]string), *routerConfig}}
	catalog, err := client.GetCatalog()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(catalog).NotTo(BeNil())
	g.Expect(len(catalog.Services)).To(Equal(1))
	g.Expect(catalog.Services[0].Name).To(Equal("name"))

}

func TestGetCatalogWithoutUpstreamServer(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := NewHandlerStubWithFunc(http.StatusOK, func(body []byte) []byte {
		return []byte(`{
  "services": {}
}`)
	})
	server, routerConfig := injectClientStub(handlerStub)
	defer server.Close()
	client := OsbClient{&RouterRestClient{routerConfig.HttpClientFactory(&http.Transport{}), make(map[string][]string), *routerConfig}}
	_, err := client.GetCatalog()

	g.Expect(err).To(HaveOccurred())
}

func TestGetCatalogWithInvalidCatalog(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := NewHandlerStubWithFunc(http.StatusOK, func(body []byte) []byte {
		return []byte("")
	})
	server, routerConfig := injectClientStub(handlerStub)
	defer server.Close()
	client := OsbClient{&RouterRestClient{routerConfig.HttpClientFactory(&http.Transport{}), make(map[string][]string), *routerConfig}}
	_, err := client.GetCatalog()

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("Can't unmarshal response from"))
}

func TestGetCatalogWithBadRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := NewHandlerStubWithFunc(http.StatusBadRequest, func(body []byte) []byte {
		return []byte(`{ "error" : "bad request"}`)
	})
	server, routerConfig := injectClientStub(handlerStub)
	defer server.Close()
	client := OsbClient{&RouterRestClient{routerConfig.HttpClientFactory(&http.Transport{}), make(map[string][]string), *routerConfig}}
	_, err := client.GetCatalog()

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(Equal("bad request"))
}
