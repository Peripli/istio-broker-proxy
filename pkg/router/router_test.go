package router

import (
	"bytes"
	"encoding/json"
	. "github.com/onsi/gomega"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/profiles"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strconv"
	"testing"
)

func TestInvalidUpdateCredentials(t *testing.T) {
	g := NewGomegaWithT(t)
	router := SetupRouter(NewProducerInterceptor(ProducerConfig{ProviderId: "x"}, DefaultPort), RouterConfig{})

	emptyBody := bytes.NewReader([]byte("{}"))
	request, _ := http.NewRequest(http.MethodPut, "https://blablub.org/v2/service_instances/134567/service_bindings/76543210/adapt_credentials", emptyBody)
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

func TestValidUpdateCredentials(t *testing.T) {
	g := NewGomegaWithT(t)
	router := SetupRouter(NewProducerInterceptor(ProducerConfig{ProviderId: "x"}, DefaultPort), RouterConfig{})

	emptyBody := bytes.NewReader([]byte(validUpdateCredentialsRequest))
	request, _ := http.NewRequest(http.MethodPut, "/v2/service_instances/1234-4567/service_bindings/7654-3210/adapt_credentials", emptyBody)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
	code := response.Code

	g.Expect(code).To(Equal(200))
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

		got := createNewUrl(externalURL, request)

		want := externalURL + "/" + helloPath
		g.Expect(got).To(Equal(want))
	})

	t.Run("Test rewrite host with parameter", func(t *testing.T) {
		g := NewGomegaWithT(t)
		body := []byte{'{', '}'}
		request, _ := http.NewRequest(http.MethodGet, "https://"+internalHost+"/"+helloPath+"?debug=true", bytes.NewReader(body))
		request.Header = make(http.Header)
		request.Header["accept"] = []string{"application/json"}

		got := createNewUrl(externalURL, request)

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
		router := SetupRouter(&NoOpInterceptor{}, RouterConfig{ForwardURL: "https://httpbin.org"})
		router.ServeHTTP(response, request)

		var bodyData struct {
			Headers map[string]string `json:"headers"`
		}

		err := json.NewDecoder(response.Body).Decode(&bodyData)
		g.Expect(err).NotTo(HaveOccurred(), "error while decoding body: %v ", response.Body)

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
		router := SetupRouter(&NoOpInterceptor{}, RouterConfig{ForwardURL: "https://httpbin.org"})
		router.ServeHTTP(response, request)

		var bodyData struct {
			Json map[string]string `json:"json"`
		}

		err := json.NewDecoder(response.Body).Decode(&bodyData)
		g.Expect(err).NotTo(HaveOccurred(), "error while decoding body: %v ", response.Body)

		got := bodyData.Json["service_id"]

		want := "6db542eb-8187-4afc-8a85-e08b4a3cc24e"
		g.Expect(got).To(Equal(want))
	})

}

func TestBadGateway(t *testing.T) {
	g := NewGomegaWithT(t)
	router := SetupRouter(&NoOpInterceptor{}, RouterConfig{ForwardURL: "doesntexist.org"})

	body := []byte{'{', '}'}
	request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/get", bytes.NewReader(body))
	request.Header = make(http.Header)
	request.Header["accept"] = []string{"application/json"}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	g.Expect(response.Code).To(Equal(502))
}

func TestAdaptCredentials(t *testing.T) {
	g := NewGomegaWithT(t)
	router := SetupRouter(NewProducerInterceptor(ProducerConfig{ProviderId: "x"}, DefaultPort), RouterConfig{})

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
	request, _ := http.NewRequest(http.MethodPut, "https://blahblubs.org/v2/service_instances/1234-4567/service_bindings/7654-3210/adapt_credentials", bytes.NewReader(body))
	request.Header = make(http.Header)
	request.Header["accept"] = []string{"application/json"}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	g.Expect(response.Code).To(Equal(200))
	g.Expect(response.Body).To(ContainSubstring(`"endpoints":[{"host":"appnethost","port":9876}]`))
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
	handlerStub := NewHandlerStub(http.StatusOK, body)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodPut, "https://blahblubs.org/v2/service_instances/123/service_bindings/456", bytes.NewReader(body))
	response := httptest.NewRecorder()
	router := SetupRouter(NewProducerInterceptor(ProducerConfig{}, DefaultPort), *routerConfig)
	router.ServeHTTP(response, request)

	var bodyData struct {
		Endpoints []interface{} `json:"endpoints"`
	}

	err := json.NewDecoder(response.Body).Decode(&bodyData)
	g.Expect(err).NotTo(HaveOccurred(), "error while decoding body: %v ", response.Body)
	g.Expect(bodyData.Endpoints).To(HaveLen(1))
}

func TestAddIstioNetworkDataProvidesEndpointHostsBasedOnSystemDomainServiceIdAndEndpointIndex(t *testing.T) {
	producerConfig := ProducerConfig{SystemDomain: "istio.sapcloud.io", ProviderId: "your-provider"}
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
	handlerStub := NewHandlerStub(http.StatusOK, body)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodPut, "https://blahblubs.org/v2/service_instances/123/service_bindings/456", bytes.NewReader(body))
	response := httptest.NewRecorder()
	router := SetupRouter(NewProducerInterceptor(producerConfig, DefaultPort), *routerConfig)
	router.ServeHTTP(response, request)

	bodyString := response.Body.String()
	expectedLength, _ := strconv.Atoi(response.Header().Get("content-length"))
	g.Expect(len(bodyString)).To(Equal(expectedLength))
	g.Expect(bodyString).To(ContainSubstring("network_data"))
	g.Expect(bodyString).To(ContainSubstring("istio.sapcloud.io"))
	g.Expect(bodyString).To(ContainSubstring("your-provider"))
	g.Expect(bodyString).To(ContainSubstring("9000"))
}

func TestIstioConfigFilesAreWritten(t *testing.T) {
	producerConfig := ProducerConfig{SystemDomain: "services.cf.dev99.sc6.istio.sapcloud.io", ProviderId: "your-provider"}
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
	handlerStub := NewHandlerStub(http.StatusOK, responseBody)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodPut, "https://blahblubs.org/v2/service_instances/123/service_bindings/456", bytes.NewReader(requestBody))
	response := httptest.NewRecorder()
	router := SetupRouter(NewProducerInterceptor(producerConfig, DefaultPort), *routerConfig)
	router.ServeHTTP(response, request)

	file, err := os.Open(path.Join(os.TempDir(), "456.yml"))
	g.Expect(err).NotTo(HaveOccurred())
	content, err := ioutil.ReadAll(file)
	contentAsString := string(content)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(contentAsString).To(ContainSubstring("147"))
	g.Expect(contentAsString).To(MatchRegexp("number: 9000"))
	g.Expect(contentAsString).To(MatchRegexp("0.456.services.cf.dev99.sc6.istio.sapcloud.io"))
}

func TestIstioConfigFilesAreNotWritable(t *testing.T) {
	producerConfig := ProducerConfig{
		SystemDomain:   "services.cf.dev99.sc6.istio.sapcloud.io",
		ProviderId:     "your-provider",
		IstioDirectory: "/non-existing-directory",
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
	handlerStub := NewHandlerStub(http.StatusOK, responseBody)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodPut, "https://blahblubs.org/v2/service_instances/123/service_bindings/error", bytes.NewReader(requestBody))
	response := httptest.NewRecorder()
	router := SetupRouter(NewProducerInterceptor(producerConfig, DefaultPort), *routerConfig)
	router.ServeHTTP(response, request)
	g.Expect(response.Code).To(Equal(500))
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
	handlerStub := NewHandlerStub(http.StatusNotFound, []byte{})
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodPut, "https://blahblubs.org/v2/service_instances/123/service_bindings/456", bytes.NewReader(body))
	response := httptest.NewRecorder()
	router := SetupRouter(&NoOpInterceptor{}, *routerConfig)
	router.ServeHTTP(response, request)

	bodyString := response.Body.String()
	g.Expect(response.Code).To(Equal(http.StatusNotFound))
	g.Expect(bodyString).To(Equal(""))
}

func TestRequestServiceBindingAddsNetworkDataToRequestIfConsumer(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{
					"credentials":
					{
 						"hostname": "10.11.241.0",
 						"port": "47637",
						"uri": "postgres://mma4G8N0isoxe17v:redacted@10.11.241.0:47637/yLO2WoE0-mCcEppn"
 					}
					}`)
	handlerStub := NewHandlerStub(http.StatusOK, body)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodPut, "https://blahblubs.org/v2/service_instances/123/service_bindings/456", bytes.NewReader(body))
	response := httptest.NewRecorder()
	router := SetupRouter(NewConsumerInterceptor(ConsumerConfig{ConsumerId: "your-consumer"}), *routerConfig)
	router.ServeHTTP(response, request)

	bodyString := handlerStub.spy.body
	g.Expect(bodyString).To(ContainSubstring("network_data"))
	g.Expect(bodyString).To(ContainSubstring(profiles.NetworkProfile))
	g.Expect(bodyString).To(ContainSubstring("consumer_id"))
}

func TestErrorCodeOfForwardIsReturned(t *testing.T) {
	g := NewGomegaWithT(t)
	handlerStub := NewHandlerStub(http.StatusServiceUnavailable, nil)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	body := []byte{'{', '}'}
	request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/status/503", bytes.NewReader(body))

	response := httptest.NewRecorder()
	router := SetupRouter(&NoOpInterceptor{}, *routerConfig)
	router.ServeHTTP(response, request)

	g.Expect(response.Code).To(Equal(503))
}

func TestReturnCodeOfGet(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte{'{', '}'}
	handlerStub := NewHandlerStub(http.StatusOK, body)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/xxx", bytes.NewReader(body))

	response := httptest.NewRecorder()
	router := SetupRouter(&NoOpInterceptor{}, *routerConfig)
	router.ServeHTTP(response, request)

	g.Expect(response.Code).To(Equal(200))
}

func TestCorrectUrlForwarded(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte{'{', '}'}
	handlerStub := NewHandlerStub(http.StatusOK, body)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/somepath", bytes.NewReader(body))

	response := httptest.NewRecorder()
	router := SetupRouter(&NoOpInterceptor{}, *routerConfig)
	router.ServeHTTP(response, request)

	g.Expect(handlerStub.spy.url).To(Equal("http://xxxxx.xx/somepath"))
}

func TestCorrectRequestParamForDelete(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{}`)
	handlerStub := NewHandlerStub(http.StatusOK, body)
	server, routerConfig := injectClientStub(handlerStub)

	defer server.Close()
	request, _ := http.NewRequest(http.MethodDelete, "https://blahblubs.org/delete?plan_id=myplan", bytes.NewReader(body))

	response := httptest.NewRecorder()
	router := SetupRouter(&NoOpInterceptor{}, *routerConfig)
	router.ServeHTTP(response, request)

	g.Expect(handlerStub.spy.url).To(Equal("http://xxxxx.xx/delete?plan_id=myplan"))
}
