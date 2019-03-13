package router

import (
	"bytes"
	"encoding/json"
	"github.com/Peripli/istio-broker-proxy/pkg/model"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	. "github.com/onsi/gomega"
)

func TestHealthEndpoint(t *testing.T) {
	g := NewGomegaWithT(t)
	router := SetupRouter(noOpInterceptor{}, Config{})

	emptyBody := bytes.NewReader([]byte(""))
	request, _ := http.NewRequest(http.MethodGet, "https://blablub.org/health", emptyBody)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
	code := response.Code

	g.Expect(code).To(Equal(http.StatusOK))
}

func TestInvalidUpdateCredentials(t *testing.T) {
	g := NewGomegaWithT(t)
	router := SetupRouter(ProducerInterceptor{ProviderID: "x"}, Config{})

	emptyBody := bytes.NewReader([]byte("{}"))
	request, _ := http.NewRequest(http.MethodPost, "https://blablub.org/v2/service_instances/134567/service_bindings/76543210/adapt_credentials", emptyBody)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
	code := response.Code

	g.Expect(code).To(Equal(http.StatusBadRequest))
}

const validUpdateCredentialsRequest = `{
    "credentials": {
	    "dbname": "dbname",
        "hostname": "mydb",
        "password": "pass",
        "port": "8080",
        "uri": "postgres://user:pass@mydb:8080/dbname",
        "username": "user"
     },
    "endpoint_mappings": [{
        "source": {"host": "mydb", "port": 8080},
        "target": {"host": "10.11.241.0", "port": 80}
	}]
}`

func TestDoNotSkipVerifyTLSIfNotConfigured(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := newHandlerStub(200, []byte(`{}`))
	_, routerConfig := injectClientStub(handlerStub)
	routerConfig.SkipVerifyTLS = true
	SetupRouter(ProducerInterceptor{ProviderID: "x"}, *routerConfig)

	skipVerify := handlerStub.spy.tr.TLSClientConfig.InsecureSkipVerify

	g.Expect(skipVerify).To(BeTrue())
}

func TestSkipVerifyTLSIfConfigured(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := newHandlerStub(200, []byte(`{}`))
	_, routerConfig := injectClientStub(handlerStub)
	routerConfig.SkipVerifyTLS = false
	SetupRouter(ProducerInterceptor{ProviderID: "x"}, *routerConfig)

	skipVerify := handlerStub.spy.tr.TLSClientConfig.InsecureSkipVerify

	g.Expect(skipVerify).To(BeFalse())
}

func TestValidUpdateCredentials(t *testing.T) {
	g := NewGomegaWithT(t)
	router := SetupRouter(ProducerInterceptor{ProviderID: "x"}, Config{})

	emptyBody := bytes.NewReader([]byte(validUpdateCredentialsRequest))
	request, _ := http.NewRequest(http.MethodPost, "/v2/service_instances/1234-4567/service_bindings/7654-3210/adapt_credentials", emptyBody)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
	code := response.Code

	g.Expect(code).To(Equal(200))
}

func TestConsumerForwardsAdpotCredentials(t *testing.T) {
	g := NewGomegaWithT(t)
	handlerStub := newHandlerStub(499, []byte(`{"error" : "abc"}`))
	server, routerConfig := injectClientStub(handlerStub)
	defer server.Close()
	router := SetupRouter(ConsumerInterceptor{ConsumerID: "x"}, *routerConfig)

	emptyBody := bytes.NewReader([]byte(validUpdateCredentialsRequest))
	request, _ := http.NewRequest(http.MethodPost, "/v2/service_instances/1234-4567/service_bindings/7654-3210/adapt_credentials", emptyBody)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
	code := response.Code

	g.Expect(code).To(Equal(499))
	err := model.HTTPErrorFromResponse(response.Code, response.Body.Bytes(), "", "")
	g.Expect(err.Error()).To(Equal("error: 'abc', description: ': from call to  '"))

}

func TestCreateNewURL(t *testing.T) {
	const internalHost = "internal-name.test"
	const externalURL = "https://external-name.test/cf"
	const helloPath = "hello"

	t.Run("Test rewrite host", func(t *testing.T) {
		g := NewGomegaWithT(t)
		body := []byte{'{', '}'}
		request, _ := http.NewRequest(http.MethodGet, "https://"+internalHost+"/"+helloPath, bytes.NewReader(body))
		request.Header = make(http.Header)
		request.Header["accept"] = []string{"application/json"}

		got := createNewURL(externalURL, request)

		want := externalURL + "/" + helloPath
		g.Expect(got).To(Equal(want))
	})

	t.Run("Test rewrite host with parameter", func(t *testing.T) {
		g := NewGomegaWithT(t)
		body := []byte{'{', '}'}
		request, _ := http.NewRequest(http.MethodGet, "https://"+internalHost+"/"+helloPath+"?debug=true", bytes.NewReader(body))
		request.Header = make(http.Header)
		request.Header["accept"] = []string{"application/json"}

		got := createNewURL(externalURL, request)

		want := externalURL + "/" + helloPath + "?debug=true"
		g.Expect(got).To(Equal(want))
	})
}

func TestRedirect(t *testing.T) {

	t.Run("Check that headers are forwarded", func(t *testing.T) {
		const testHeaderKey = "X-Broker-Api-Version"
		const testHeaderValue = "2.13"
		g := NewGomegaWithT(t)

		body := []byte{'{', '}'}
		request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/headers", bytes.NewReader(body))
		request.Header = make(http.Header)
		request.Header.Set("accept", "application/json")
		request.Header.Set(testHeaderKey, testHeaderValue)

		response := httptest.NewRecorder()
		router := SetupRouter(&noOpInterceptor{}, Config{ForwardURL: "https://httpbin.org"})
		router.ServeHTTP(response, request)

		var bodyData struct {
			Headers map[string]string `json:"headers"`
		}

		err := json.NewDecoder(response.Body).Decode(&bodyData)
		g.Expect(err).NotTo(HaveOccurred(), "error while decoding")

		got := bodyData.Headers[testHeaderKey]

		want := request.Header.Get(testHeaderKey)
		g.Expect(got).To(Equal(want))
	})

	t.Run("Check that the request body is forwarded for PUT", func(t *testing.T) {
		g := NewGomegaWithT(t)
		body := []byte(`{"service_id":"6db542eb-8187-4afc-8a85-e08b4a3cc24e","plan_id":"c3320e0f-5866-4f14-895e-48bc92a4245c"}`)
		request, _ := http.NewRequest(http.MethodPut, "https://blahblubs.org/put", bytes.NewReader(body))
		request.Header = make(http.Header)
		request.Header.Set("accept", "application/json")
		request.Header.Set("'Content-Type", "application/json")

		response := httptest.NewRecorder()
		router := SetupRouter(&noOpInterceptor{}, Config{ForwardURL: "https://httpbin.org"})
		router.ServeHTTP(response, request)

		var bodyData struct {
			JSON map[string]string `json:"json"`
		}

		err := json.NewDecoder(response.Body).Decode(&bodyData)
		g.Expect(err).NotTo(HaveOccurred(), "error while decoding body: %v ", response.Body)

		got := bodyData.JSON["service_id"]

		want := "6db542eb-8187-4afc-8a85-e08b4a3cc24e"
		g.Expect(got).To(Equal(want))
	})

}

func TestBadGateway(t *testing.T) {
	g := NewGomegaWithT(t)
	router := SetupRouter(&noOpInterceptor{}, Config{ForwardURL: "doesntexist.org"})

	body := []byte{'{', '}'}
	request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/get", bytes.NewReader(body))
	request.Header = make(http.Header)
	request.Header["accept"] = []string{"application/json"}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	g.Expect(response.Code).To(Equal(http.StatusBadGateway))
	err := model.HTTPErrorFromResponse(response.Code, response.Body.Bytes(), "", "")
	g.Expect(err.(*model.HTTPError).Description).To(ContainSubstring(`Get doesntexist.org/get: unsupported protocol scheme ""`))
}

func TestAdaptCredentials(t *testing.T) {
	g := NewGomegaWithT(t)
	router := SetupRouter(ProducerInterceptor{ProviderID: "x"}, Config{})

	body := []byte(`{
"credentials": {
 "dbname": "yLO2WoE0-mCcEppn",
 "hostname": "10.11.241.0",
 "password": "redacted",
 "port": "47637",
 "ports": {
  "5432/tcp": "47637"
 },
 "uri": "postgres://mma4G8N0isoxe17v:redacted@10.11.241.0:47637/yLO2WoE0-mCcEppn",
 "username": "mma4G8N0isoxe17v"
},
"endpoint_mappings": [{
  "source": {"host": "10.11.241.0", "port": 47637},
  "target": {"host": "appnethost", "port": 9876}
	}]
}
`)
	request, _ := http.NewRequest(http.MethodPost, "https://blahblubs.org/v2/service_instances/1234-4567/service_bindings/7654-3210/adapt_credentials", bytes.NewReader(body))
	request.Header = make(http.Header)
	request.Header["accept"] = []string{"application/json"}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	g.Expect(response.Code).To(Equal(200))
	g.Expect(response.Body).To(ContainSubstring(`"endpoints":[{"host":"appnethost","port":9876}]`))
	g.Expect(response.Body).To(ContainSubstring(`"hostname":"appnethost"`))
}

func TestCreateServiceBindingContainsEndpoints(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{
					"credentials":
					{
 						"hostname": "10.11.241.0",
 						"port": "47637",
                        "end_points": [
                        {
                            "host": "10.11.241.0",
                            "port": 47637
                        }],
						"uri": "postgres://mma4G8N0isoxe17v:redacted@10.11.241.0:47637/yLO2WoE0-mCcEppn"
 					}
					}`)
	requestBody := []byte(`{
					"network_data":
					{
                        "data":
                        {
                            "consumer_id": "147"
                        }
 					}
					}`)
	handlerStub := newHandlerStub(http.StatusOK, body)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodPut, "https://blahblubs.org/v2/service_instances/123/service_bindings/456", bytes.NewReader(requestBody))
	response := httptest.NewRecorder()
	router := SetupRouter(ProducerInterceptor{IstioDirectory: os.TempDir(), NetworkProfile: "urn:local.test:public"}, *routerConfig)
	router.ServeHTTP(response, request)
	g.Expect(response.Code).To(Equal(http.StatusOK))

	var bodyData struct {
		Endpoints []interface{} `json:"endpoints"`
	}

	err := json.NewDecoder(response.Body).Decode(&bodyData)
	g.Expect(err).NotTo(HaveOccurred(), "error while decoding body: %v ", response.Body)
	g.Expect(bodyData.Endpoints).To(HaveLen(1))
}

func TestBindWithoutConsumerId(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{
					"credentials":
					{
 						"hostname": "10.11.241.0",
 						"port": "47637",
                        "end_points": [
                        {
                            "host": "10.11.241.0",
                            "port": 47637
                        }],
						"uri": "postgres://mma4G8N0isoxe17v:redacted@10.11.241.0:47637/yLO2WoE0-mCcEppn"
 					}
					}`)
	handlerStub := newHandlerStub(http.StatusOK, body)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodPut, "https://blahblubs.org/v2/service_instances/123/service_bindings/456", bytes.NewReader(body))
	response := httptest.NewRecorder()
	router := SetupRouter(ProducerInterceptor{IstioDirectory: os.TempDir(), NetworkProfile: "urn:local.test:public"}, *routerConfig)
	router.ServeHTTP(response, request)
	g.Expect(response.Code).To(Equal(http.StatusBadRequest))
	var err model.HTTPError
	json.Unmarshal(response.Body.Bytes(), &err)
	g.Expect(err.Description).To(Equal("no consumer ID included in bind request"))
	g.Expect(err.ErrorMsg).To(Equal("InvalidConsumerID"))

}

func TestAddIstioNetworkDataProvidesEndpointHostsBasedOnSystemDomainServiceIdAndEndpointIndex(t *testing.T) {
	producerConfig := ProducerInterceptor{
		SystemDomain:     "my.arbitrary.domain.io",
		ProviderID:       "your-provider",
		LoadBalancerPort: 9000,
		IstioDirectory:   os.TempDir(),
		NetworkProfile:   "urn:local.test:public"}
	g := NewGomegaWithT(t)
	body := []byte(`{
					"credentials":
					{
 						"hostname": "10.11.241.0",
 						"port": "47637",
                        "end_points": [
                        {
                            "host": "10.11.241.0",
                            "port": 47637
                        }],
						"uri": "postgres://mma4G8N0isoxe17v:redacted@10.11.241.0:47637/yLO2WoE0-mCcEppn"
 					}
					}`)
	requestBody := []byte(`{
					"network_data":
					{
                        "data":
                        {
                            "consumer_id": "147"
                        }
 					}
					}`)
	handlerStub := newHandlerStub(http.StatusOK, body)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodPut, "https://blahblubs.org/v2/service_instances/123/service_bindings/456", bytes.NewReader(requestBody))
	response := httptest.NewRecorder()
	router := SetupRouter(producerConfig, *routerConfig)
	router.ServeHTTP(response, request)
	g.Expect(response.Code).To(Equal(http.StatusOK))

	bodyString := response.Body.String()
	g.Expect(bodyString).To(ContainSubstring("network_data"))
	g.Expect(bodyString).To(ContainSubstring("my.arbitrary.domain.io"))
	g.Expect(bodyString).To(ContainSubstring("your-provider"))
	g.Expect(bodyString).To(ContainSubstring("9000"))
}

func TestIstioConfigFilesAreWritten(t *testing.T) {
	producerInterceptor := ProducerInterceptor{
		SystemDomain:   "services.cf.dev99.sc6.my.arbitrary.domain.io",
		ProviderID:     "your-provider",
		IstioDirectory: os.TempDir(),
		NetworkProfile: "urn:local.test:public"}
	g := NewGomegaWithT(t)
	responseBody := []byte(`{
					"credentials":
					{
 						"hostname": "10.11.241.0",
 						"port": "47637",
                        "end_points": [
                        {
                            "host": "10.11.241.0",
                            "port": 47637
                        }],
						"uri": "postgres://mma4G8N0isoxe17v:redacted@10.11.241.0:47637/yLO2WoE0-mCcEppn"
 					}
					}`)
	requestBody := []byte(`{
					"network_data":
					{
                        "data":
                        {
                            "consumer_id": "147"
                        }
 					}
					}`)
	handlerStub := newHandlerStub(http.StatusOK, responseBody)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodPut, "https://blahblubs.org/v2/service_instances/123/service_bindings/456", bytes.NewReader(requestBody))
	response := httptest.NewRecorder()
	router := SetupRouter(producerInterceptor, *routerConfig)
	router.ServeHTTP(response, request)

	file, err := os.Open(path.Join(producerInterceptor.IstioDirectory, "456.yml"))
	g.Expect(err).NotTo(HaveOccurred())
	content, err := ioutil.ReadAll(file)
	contentAsString := string(content)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(contentAsString).To(ContainSubstring("147"))
	g.Expect(contentAsString).To(MatchRegexp("number: 9000"))
	g.Expect(contentAsString).To(MatchRegexp("0.456.services.cf.dev99.sc6.my.arbitrary.domain.io"))
	g.Expect(contentAsString).To(MatchRegexp("/etc/istio/certs/cf-service.crt"))
}

func TestIstioConfigFilesAreNotWritable(t *testing.T) {
	producerConfig := ProducerInterceptor{
		SystemDomain:   "services.cf.dev99.sc6.my.arbitrary.domain.io",
		ProviderID:     "your-provider",
		IstioDirectory: "/non-existing-directory",
		NetworkProfile: "urn:local.test:public",
	}
	g := NewGomegaWithT(t)
	responseBody := []byte(`{
					"credentials":
					{
 						"hostname": "10.11.241.0",
 						"port": "47637",
                        "end_points": [
                        {
                            "host": "10.11.241.0",
                            "port": 47637
                        }],
						"uri": "postgres://mma4G8N0isoxe17v:redacted@10.11.241.0:47637/yLO2WoE0-mCcEppn"
 					}
					}`)
	requestBody := []byte(`{
					"network_data":
					{
                        "data":
                        {
                            "consumer_id": "147"
                        }
 					}
					}`)
	handlerStub := newHandlerStub(http.StatusOK, responseBody)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodPut, "https://blahblubs.org/v2/service_instances/123/service_bindings/error", bytes.NewReader(requestBody))
	response := httptest.NewRecorder()
	router := SetupRouter(producerConfig, *routerConfig)
	router.ServeHTTP(response, request)
	g.Expect(response.Code).To(Equal(500))
	err := model.HTTPErrorFromResponse(response.Code, response.Body.Bytes(), "", "")
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.(*model.HTTPError).Description).To(ContainSubstring("Unable to write istio configuration to file"))
}

func TestBindWithInvalidRequest(t *testing.T) {
	producerConfig := ProducerInterceptor{}
	g := NewGomegaWithT(t)
	handlerStub := newHandlerStub(http.StatusOK, []byte(``))
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodPut, "https://blahblubs.org/v2/service_instances/123/service_bindings/error", bytes.NewReader([]byte(`[]`)))
	response := httptest.NewRecorder()
	router := SetupRouter(producerConfig, *routerConfig)
	router.ServeHTTP(response, request)
	g.Expect(response.Code).To(Equal(http.StatusBadRequest))
	err := model.HTTPErrorFromResponse(response.Code, response.Body.Bytes(), "", "")
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.(*model.HTTPError).Description).To(ContainSubstring("cannot unmarshal array into Go value"))
}

func TestHttpClientError(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{
					"credentials":
					{
                        "hostname": "10.11.241.0",
                        "port": "47637",
                        "end_points": [
                        {
                            "host": "10.11.241.0",
                            "port": 47637
                        }],
                        "uri": "postgres://mma4G8N0isoxe17v:redacted@10.11.241.0:47637/yLO2WoE0-mCcEppn"
                    }
                    }`)
	handlerStub := newHandlerStub(http.StatusNotFound, []byte(`{ "error": "Not found", "description": "Unable to find entry"}`))
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodPut, "https://blahblubs.org/v2/service_instances/123/service_bindings/456", bytes.NewReader(body))
	response := httptest.NewRecorder()
	router := SetupRouter(&noOpInterceptor{}, *routerConfig)
	router.ServeHTTP(response, request)

	g.Expect(response.Code).To(Equal(http.StatusNotFound))
	err := model.HTTPErrorFromResponse(response.Code, response.Body.Bytes(), "", "")
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.(*model.HTTPError).ErrorMsg).To(Equal("Not found"))
	g.Expect(err.(*model.HTTPError).Description).To(ContainSubstring("Unable to find entry"))
}

func TestRequestServiceBindingAddsNetworkDataToRequestIfConsumer(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{
					"credentials":
					{
						"hostname": "10.11.241.0",
						"port": "47637",
						"uri": "postgres://mma4G8N0isoxe17v:redacted@10.11.241.0:47637/yLO2WoE0-mCcEppn"
					},
					"network_data":
					{
						"network_profile_id": "network-profile"
					}
					}`)

	handlerStub := newHandlerStub(http.StatusOK, body)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodPut, "https://blahblubs.org/v2/service_instances/123/service_bindings/456", bytes.NewReader(body))
	response := httptest.NewRecorder()
	router := SetupRouter(ConsumerInterceptor{ConsumerID: "your-consumer", NetworkProfile: "network-profile"}, *routerConfig)
	router.ServeHTTP(response, request)

	g.Expect(len(handlerStub.spy.body)).To(Equal(2))
	bodyString := handlerStub.spy.body[0]
	g.Expect(bodyString).To(ContainSubstring("network_data"))
	g.Expect(bodyString).To(ContainSubstring("network-profile"))
	g.Expect(bodyString).To(ContainSubstring("consumer_id"))
}

func TestRequestServiceBindingAddsNetworkDataToRequestIfConsumerForBrokerAPI(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{
					"network_data":
					{
						"network_profile_id": "my-network-profile"
					}
					}`)

	handlerStub := newHandlerStub(http.StatusOK, body)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodPut, "https://blahblubs.org/v1/osb/2324552-34535345-34534535/v2/service_instances/123/service_bindings/456", bytes.NewReader(body))
	response := httptest.NewRecorder()
	router := SetupRouter(ConsumerInterceptor{ConsumerID: "your-consumer", NetworkProfile: "my-network-profile"}, *routerConfig)
	router.ServeHTTP(response, request)

	g.Expect(len(handlerStub.spy.body)).To(Equal(2))
	bodyString := handlerStub.spy.body[0]
	g.Expect(bodyString).To(ContainSubstring("consumer_id"))
}

func TestErrorCodeOfForwardIsReturned(t *testing.T) {
	g := NewGomegaWithT(t)
	handlerStub := newHandlerStub(http.StatusServiceUnavailable, []byte(`{ "error": "xxx", "description": "yyy"}`))
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/status/503", bytes.NewReader([]byte("{}")))

	response := httptest.NewRecorder()
	router := SetupRouter(&noOpInterceptor{}, *routerConfig)
	router.ServeHTTP(response, request)

	g.Expect(response.Code).To(Equal(503))
	err := model.HTTPErrorFromResponse(response.Code, response.Body.Bytes(), "", "")
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.(*model.HTTPError).ErrorMsg).To(Equal("xxx"))
	g.Expect(err.(*model.HTTPError).Description).To(ContainSubstring("yyy"))
}

func TestReturnCodeOfGet(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte{'{', '}'}
	handlerStub := newHandlerStub(http.StatusOK, body)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/xxx", bytes.NewReader(body))

	response := httptest.NewRecorder()
	router := SetupRouter(&noOpInterceptor{}, *routerConfig)
	router.ServeHTTP(response, request)

	g.Expect(response.Code).To(Equal(200))
}

func TestCorrectUrlForwarded(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte{'{', '}'}
	handlerStub := newHandlerStub(http.StatusOK, body)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/somepath", bytes.NewReader(body))

	response := httptest.NewRecorder()
	router := SetupRouter(&noOpInterceptor{}, *routerConfig)
	router.ServeHTTP(response, request)

	g.Expect(handlerStub.spy.url).To(Equal("http://xxxxx.xx/somepath"))
}

func TestDeleteBinding(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte{'{', '}'}
	handlerStub := newHandlerStub(http.StatusOK, body)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodDelete, "https://blahblubs.org/v2/service_instances/123/service_bindings/456?parameter=true", bytes.NewReader(body))

	response := httptest.NewRecorder()
	var bindID = ""
	router := SetupRouter(&DeleteInterceptor{deleteCallback: func(innerBindId string) {
		bindID = innerBindId
	}}, *routerConfig)
	router.ServeHTTP(response, request)

	g.Expect(bindID).To(Equal("456"))
	g.Expect(response.Code).To(Equal(http.StatusOK))
	g.Expect(handlerStub.spy.url).To(ContainSubstring("parameter=true"))
}

func TestDeleteBindingForBrokerAPI(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte{'{', '}'}
	handlerStub := newHandlerStub(http.StatusOK, body)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodDelete, "https://blahblubs.org/v1/osb/23234234-324234234-234234/v2/service_instances/123/service_bindings/321?parameter=false", bytes.NewReader(body))

	response := httptest.NewRecorder()
	var bindID = ""
	router := SetupRouter(&DeleteInterceptor{deleteCallback: func(innerBindId string) {
		bindID = innerBindId
	}}, *routerConfig)
	router.ServeHTTP(response, request)

	g.Expect(bindID).To(Equal("321"))
	g.Expect(response.Code).To(Equal(http.StatusOK))
	g.Expect(handlerStub.spy.url).To(ContainSubstring("parameter=false"))
}

func TestDeleteBindingNotFound(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte{'{', '}'}
	handlerStub := newHandlerStub(http.StatusNotFound, body)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodDelete, "https://blahblubs.org/v2/service_instances/123/service_bindings/456", bytes.NewReader(body))

	response := httptest.NewRecorder()
	router := SetupRouter(&noOpInterceptor{}, *routerConfig)
	router.ServeHTTP(response, request)

	g.Expect(response.Code).To(Equal(http.StatusNotFound))
}

func TestForwardGetCatalog(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{"services": [{ "name" : "abc", "plans":[{}] } ] }`)
	handlerStub := newHandlerStub(http.StatusOK, body)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/v2/catalog", bytes.NewReader(make([]byte, 0)))

	response := httptest.NewRecorder()
	router := SetupRouter(&ProducerInterceptor{ServiceNamePrefix: "istio-", PlanMetaData: "{}"}, *routerConfig)
	router.ServeHTTP(response, request)
	responseBody := response.Body.String()
	g.Expect(responseBody).To(ContainSubstring("istio-abc"))
}

func TestForwardGetCatalogForBrokerAPI(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{"services": [{ "name" : "name", "plans":[{}] } ] }`)
	handlerStub := newHandlerStub(http.StatusOK, body)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/v1/osb/23-34534534-453/v2/catalog", bytes.NewReader(make([]byte, 0)))

	response := httptest.NewRecorder()
	router := SetupRouter(&ProducerInterceptor{ServiceNamePrefix: "prefix-", PlanMetaData: "{}"}, *routerConfig)
	router.ServeHTTP(response, request)
	responseBody := response.Body.String()
	g.Expect(responseBody).To(ContainSubstring("prefix-name"))
}

func TestCorrectRequestParamForDelete(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{}`)
	handlerStub := newHandlerStub(http.StatusOK, body)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()
	request, _ := http.NewRequest(http.MethodDelete, "https://blahblubs.org/delete?plan_id=myplan", bytes.NewReader(body))

	response := httptest.NewRecorder()
	router := SetupRouter(&noOpInterceptor{}, *routerConfig)
	router.ServeHTTP(response, request)

	g.Expect(handlerStub.spy.url).To(Equal("http://xxxxx.xx/delete?plan_id=myplan"))
}

type DeleteInterceptor struct {
	noOpInterceptor
	deleteCallback func(bindID string)
}

func (c DeleteInterceptor) PostDelete(bindID string) {
	c.deleteCallback(bindID)
}
