package main

import (
	"bytes"
	"encoding/json"
	. "github.com/onsi/gomega"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/endpoints"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestInvalidUpdateCredentials(t *testing.T) {
	g := NewGomegaWithT(t)
	router := setupRouter()

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
	router := setupRouter()

	emptyBody := bytes.NewReader([]byte(validUpdateCredentialsRequest))
	request, _ := http.NewRequest(http.MethodPut, "/v2/service_instances/1234-4567/service_bindings/7654-3210/adapt_credentials", emptyBody)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
	code := response.Code

	g.Expect(code).To(Equal(200))
}

func TestDefaultPortUsed(t *testing.T) {
	g := NewGomegaWithT(t)

	readPort()

	g.Expect(config.port).To(Equal(DefaultPort))
}

func TestCustomPortUsed(t *testing.T) {
	g := NewGomegaWithT(t)
	oldPort := os.Getenv("PORT")
	defer func() {
		os.Setenv("PORT", oldPort)
	}()
	expectedPort := "1234"
	os.Setenv("PORT", expectedPort)

	readPort()

	g.Expect(config.port).To(Equal(1234))
}

func TestCreateNewURL(t *testing.T) {
	const internalHost = "internal-name.test"
	const externalURL = "https://external-name.test/cf"
	const path = "hello"

	t.Run("Test rewrite host", func(t *testing.T) {
		g := NewGomegaWithT(t)
		body := []byte{'{', '}'}
		request, _ := http.NewRequest(http.MethodGet, "https://"+internalHost+"/"+path, bytes.NewReader(body))
		request.Header = make(http.Header)
		request.Header["accept"] = []string{"application/json"}

		got := createNewUrl(externalURL, request)

		want := externalURL + "/" + path
		g.Expect(got).To(Equal(want))
	})

	t.Run("Test rewrite host with parameter", func(t *testing.T) {
		g := NewGomegaWithT(t)
		body := []byte{'{', '}'}
		request, _ := http.NewRequest(http.MethodGet, "https://"+internalHost+"/"+path+"?debug=true", bytes.NewReader(body))
		request.Header = make(http.Header)
		request.Header["accept"] = []string{"application/json"}

		got := createNewUrl(externalURL, request)

		want := externalURL + "/" + path + "?debug=true"
		g.Expect(got).To(Equal(want))
	})
}

func TestRedirect(t *testing.T) {
	config.forwardURL = "https://httpbin.org"

	t.Run("Check return code of redirected get", func(t *testing.T) {
		g := NewGomegaWithT(t)
		body := []byte{'{', '}'}
		request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/get", bytes.NewReader(body))
		request.Header = make(http.Header)
		request.Header["accept"] = []string{"application/json"}

		response := httptest.NewRecorder()
		router := setupRouter()
		router.ServeHTTP(response, request)
		got := response.Code

		want := 200
		g.Expect(got).To(Equal(want))
	})

	t.Run("Check URL in response", func(t *testing.T) {
		g := NewGomegaWithT(t)
		body := []byte{'{', '}'}
		request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/get", bytes.NewReader(body))
		request.Header = make(http.Header)
		request.Header["accept"] = []string{"application/json"}
		response := httptest.NewRecorder()
		router := setupRouter()
		router.ServeHTTP(response, request)

		var bodyData struct {
			URL string `json:"url"`
		}

		err := json.NewDecoder(response.Body).Decode(&bodyData)
		g.Expect(err).NotTo(HaveOccurred(), "error while decoding body: %v ", response.Body)

		got := bodyData.URL
		want := config.forwardURL + "/get"
		g.Expect(got).To(Equal(want))
	})

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
		router := setupRouter()
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
		router := setupRouter()
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

	t.Run("Check that the request param is forwarded for DELETE", func(t *testing.T) {
		g := NewGomegaWithT(t)
		body := []byte(`{}`)
		expectedPlan := "myplan"
		request, _ := http.NewRequest(http.MethodDelete, "https://blahblubs.org/delete?plan_id="+expectedPlan, bytes.NewReader(body))
		request.Header = make(http.Header)
		request.Header.Set("accept", "application/json")
		request.Header.Set("'Content-Type", "application/json")

		response := httptest.NewRecorder()
		router := setupRouter()
		router.ServeHTTP(response, request)

		var bodyData struct {
			Args map[string]string `json:"args"`
		}

		err := json.NewDecoder(response.Body).Decode(&bodyData)
		g.Expect(err).NotTo(HaveOccurred(), "error while decoding body: %v ", response.Body)

		g.Expect(bodyData.Args["plan_id"]).To(Equal(expectedPlan))
	})

}

func TestBadGateway(t *testing.T) {
	g := NewGomegaWithT(t)
	router := setupRouter()
	config.forwardURL = "doesntexist.org"

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
	router := setupRouter()
	config.forwardURL = ""

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

func TestEmptyBodyInTranslate(t *testing.T) {
	g := NewGomegaWithT(t)

	body, err := endpoints.GenerateEndpoint(make([]byte, 0))

	g.Expect(err).Should(HaveOccurred())
	g.Expect(body).To(BeEmpty())
}

func TestCreateServiceBindingContainsEndpoints(t *testing.T) {
	config.forwardURL = "http://xxxxx.xx"
	g := NewGomegaWithT(t)
	body := []byte(`{
					"credentials":
					{
 						"hostname": "10.11.241.0",
 						"port": "47637",
						"uri": "postgres://mma4G8N0isoxe17v:redacted@10.11.241.0:47637/yLO2WoE0-mCcEppn"
 					}
					}`)
	handlerStub := NewHandlerStub(http.StatusServiceUnavailable, body)
	server := injectClientStub(handlerStub)

	defer server.Close()

	request, _ := http.NewRequest(http.MethodPut, "https://blahblubs.org/v2/service_instances/123/service_bindings/456", bytes.NewReader(body))
	response := httptest.NewRecorder()
	router := setupRouter()
	router.ServeHTTP(response, request)

	var bodyData struct {
		Endpoints []interface{} `json:"endpoints"`
	}

	err := json.NewDecoder(response.Body).Decode(&bodyData)
	g.Expect(err).NotTo(HaveOccurred(), "error while decoding body: %v ", response.Body)
	g.Expect(bodyData.Endpoints).To(HaveLen(1))
}

func TestErrorCodeOfForwardIsReturned(t *testing.T) {
	config.forwardURL = "http://xxxxx.xx"
	g := NewGomegaWithT(t)
	handlerStub := NewHandlerStub(http.StatusServiceUnavailable, nil)
	server := injectClientStub(handlerStub)

	defer server.Close()

	body := []byte{'{', '}'}
	request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/status/503", bytes.NewReader(body))

	response := httptest.NewRecorder()
	router := setupRouter()
	router.ServeHTTP(response, request)

	g.Expect(response.Code).To(Equal(503))
}
