package router

import (
	"encoding/json"
	"github.com/Peripli/istio-broker-proxy/pkg/model"
	. "github.com/onsi/gomega"
	"net/http"
	"net/url"
	"testing"
)

func TestAdaptCredentialsWithProxy(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := newHandlerStubWithFunc(http.StatusOK, func(body []byte) []byte {
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
	client := osbClient{&restClient{routerConfig.HTTPClientFactory(&http.Transport{}), &http.Request{URL: &url.URL{}}, *routerConfig}}
	binding, err := client.adaptCredentials(model.PostgresCredentials{
		Port:     47637,
		Hostname: "10.11.241.0",
		URI:      "postgres://mma4G8N0isoxe17v:redacted@10.11.241.0:47637/yLO2WoE0-mCcEppn",
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
	g.Expect(postgresCredentials.URI).To(Equal("postgres://mma4G8N0isoxe17v:redacted@postgres.catalog.svc.cluster.local:5555/yLO2WoE0-mCcEppn"))

}

func TestAdaptCredentialsCalledWithCorrectPath(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := newHandlerStub(http.StatusOK, []byte(`{}`))
	server, routerConfig := injectClientStub(handlerStub)
	routerConfig.ForwardURL = "https://myhost"
	defer server.Close()
	client := osbClient{&restClient{routerConfig.HTTPClientFactory(&http.Transport{}), &http.Request{URL: &url.URL{Host: "original-host",
		Path: "/v2/service_instances/552c6306-fd6a-11e8-b5d9-1287e5b96b40/service_bindings/5e58a9a6-fd6a-11e8-b5d9-1287e5b96b40"}}, *routerConfig}}
	binding, err := client.adaptCredentials(model.Credentials{}, []model.EndpointMapping{{}})

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(binding).NotTo(BeNil())

	g.Expect(handlerStub.spy.url).To(Equal("https://myhost/v2/service_instances/552c6306-fd6a-11e8-b5d9-1287e5b96b40/service_bindings/5e58a9a6-fd6a-11e8-b5d9-1287e5b96b40/adapt_credentials"))
}

func TestAdaptCredentialsWithBadRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := newHandlerStub(http.StatusBadRequest, []byte(`{"error" : "myerror", "description" : "mydescription"}`))
	server, routerConfig := injectClientStub(handlerStub)
	defer server.Close()
	client := osbClient{&restClient{routerConfig.HTTPClientFactory(&http.Transport{}), &http.Request{URL: &url.URL{}}, *routerConfig}}
	_, err := client.adaptCredentials(model.PostgresCredentials{}.ToCredentials(), []model.EndpointMapping{})

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.(*model.HTTPError).StatusCode).To(Equal(http.StatusBadRequest))
	g.Expect(err.(*model.HTTPError).ErrorMsg).To(Equal("myerror"))
	g.Expect(err.(*model.HTTPError).Description).To(Equal("mydescription: from call to POST http://xxxxx.xx/adapt_credentials"))
}

func TestAdaptCredentialsWithInvalidJson(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := newHandlerStub(http.StatusOK, []byte(""))
	server, routerConfig := injectClientStub(handlerStub)
	defer server.Close()
	client := osbClient{&restClient{routerConfig.HTTPClientFactory(&http.Transport{}), &http.Request{URL: &url.URL{}}, *routerConfig}}
	_, err := client.adaptCredentials(model.PostgresCredentials{}.ToCredentials(), []model.EndpointMapping{})

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("Can't unmarshal response from"))
}

func TestGetCatalog(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := newHandlerStubWithFunc(http.StatusOK, func(body []byte) []byte {
		return []byte(`{
  "services": [{
    "id": "id",
    "name": "name" }]
}`)
	})
	server, routerConfig := injectClientStub(handlerStub)
	defer server.Close()
	client := osbClient{&restClient{routerConfig.HTTPClientFactory(&http.Transport{}), &http.Request{URL: &url.URL{}}, *routerConfig}}
	catalog, err := client.getCatalog()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(catalog).NotTo(BeNil())
	g.Expect(len(catalog.Services)).To(Equal(1))
	g.Expect(catalog.Services[0].Name).To(Equal("name"))

}

func TestGetCatalogWithoutUpstreamServer(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := newHandlerStubWithFunc(http.StatusOK, func(body []byte) []byte {
		return []byte(`{
  "services": {}
}`)
	})
	server, routerConfig := injectClientStub(handlerStub)
	defer server.Close()
	client := osbClient{&restClient{routerConfig.HTTPClientFactory(&http.Transport{}), &http.Request{URL: &url.URL{}}, *routerConfig}}
	_, err := client.getCatalog()

	g.Expect(err).To(HaveOccurred())
}

func TestGetCatalogWithInvalidCatalog(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := newHandlerStubWithFunc(http.StatusOK, func(body []byte) []byte {
		return []byte("")
	})
	server, routerConfig := injectClientStub(handlerStub)
	defer server.Close()
	client := osbClient{&restClient{routerConfig.HTTPClientFactory(&http.Transport{}), &http.Request{URL: &url.URL{}}, *routerConfig}}
	_, err := client.getCatalog()

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("Can't unmarshal response from"))
}

func TestGetCatalogWithBadRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := newHandlerStubWithFunc(http.StatusBadRequest, func(body []byte) []byte {
		return []byte(`{ "error" : "BadRequest"}`)
	})
	server, routerConfig := injectClientStub(handlerStub)
	defer server.Close()
	client := osbClient{&restClient{routerConfig.HTTPClientFactory(&http.Transport{}), &http.Request{URL: &url.URL{}}, *routerConfig}}
	_, err := client.getCatalog()

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(Equal("error: 'BadRequest', description: ': from call to GET http://xxxxx.xx'"))
}

func TestUnbind(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := newHandlerStubWithFunc(http.StatusOK, func(body []byte) []byte {
		return []byte(`{}`)
	})
	server, routerConfig := injectClientStub(handlerStub)
	defer server.Close()
	client := osbClient{&restClient{routerConfig.HTTPClientFactory(&http.Transport{}),
		&http.Request{URL: &url.URL{Host: "yyyy:123", Path: "/v2/service_instances/1/service_bindings/2", RawQuery: "query_parameter=value"}}, *routerConfig}}
	err := client.unbind()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(handlerStub.spy.url).To(Equal("http://xxxxx.xx/v2/service_instances/1/service_bindings/2?query_parameter=value"))
}
