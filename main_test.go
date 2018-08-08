package main

import (
	"bytes"
	"encoding/json"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestInitialInfo(t *testing.T) {
	g := NewGomegaWithT(t)
	request, _ := http.NewRequest(http.MethodGet, "/notused", nil)
	response := httptest.NewRecorder()

	info(response, request)
	got := response.Body.String()

	want := ""
	g.Expect(got).To(Equal(want))
}

func TestInfoAfterRequest(t *testing.T) {
	g := NewGomegaWithT(t)
	config.ForwardURL = "doesntexist.org"
	body := []byte{'{', '}'}
	request, _ := http.NewRequest(http.MethodGet, "/anything", bytes.NewReader(body))
	response := httptest.NewRecorder()
	redirect(response, request)

	infoRequest, _ := http.NewRequest(http.MethodGet, "/notused", nil)
	infoResponse := httptest.NewRecorder()
	info(infoResponse, infoRequest)

	infoResult := infoResponse.Body.String()
	g.Expect(infoResult).NotTo(BeEmpty())
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

	g.Expect(config.port).To(Equal(expectedPort))
}

func TestCreateNewURL(t *testing.T) {
	const internalHost = "internal-name.test"
	const externalURL = "external-name.test/cf"
	const path = "hello"

	t.Run("Test rewrite host", func(t *testing.T) {
		g := NewGomegaWithT(t)
		body := []byte{'{', '}'}
		request, _ := http.NewRequest(http.MethodGet, "https://"+internalHost+"/"+path, bytes.NewReader(body))
		request.Header = make(http.Header)
		request.Header["accept"] = []string{"application/json"}

		got := createNewUrl(externalURL, request)

		want := "https://" + externalURL + "/" + path
		g.Expect(got).To(Equal(want))
	})

	t.Run("Test rewrite host with parameter", func(t *testing.T) {
		g := NewGomegaWithT(t)
		body := []byte{'{', '}'}
		request, _ := http.NewRequest(http.MethodGet, "https://"+internalHost+"/"+path+"?debug=true", bytes.NewReader(body))
		request.Header = make(http.Header)
		request.Header["accept"] = []string{"application/json"}

		got := createNewUrl(externalURL, request)

		want := "https://" + externalURL + "/" + path + "?debug=true"
		g.Expect(got).To(Equal(want))
	})
}

func TestRedirect(t *testing.T) {
	config.ForwardURL = "httpbin.org"

	t.Run("Check return code of redirected get", func(t *testing.T) {
		g := NewGomegaWithT(t)
		body := []byte{'{', '}'}
		request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/get", bytes.NewReader(body))
		request.Header = make(http.Header)
		request.Header["accept"] = []string{"application/json"}

		response := httptest.NewRecorder()
		redirect(response, request)
		got := response.Code

		want := 200
		g.Expect(got).To(Equal(want))
	})

	t.Run("Check return code of redirected get with error code", func(t *testing.T) {
		g := NewGomegaWithT(t)
		body := []byte{'{', '}'}
		request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/status/503", bytes.NewReader(body))
		request.Header = make(http.Header)
		request.Header["accept"] = []string{"application/json"}

		response := httptest.NewRecorder()
		redirect(response, request)
		got := response.Code

		want := 503
		g.Expect(got).To(Equal(want))

	})

	t.Run("Check URL in response", func(t *testing.T) {
		g := NewGomegaWithT(t)
		body := []byte{'{', '}'}
		request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/get", bytes.NewReader(body))
		request.Header = make(http.Header)
		request.Header["accept"] = []string{"application/json"}
		response := httptest.NewRecorder()
		redirect(response, request)

		var bodyData struct {
			URL string `json:"url"`
		}

		err := json.NewDecoder(response.Body).Decode(&bodyData)
		g.Expect(err).NotTo(HaveOccurred(),"error while decoding body: %v ", response.Body)

		got := bodyData.URL
		want := "https://" + config.ForwardURL + "/get"
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
		redirect(response, request)

		var bodyData struct {
			Headers map[string]string `json:"headers"`
		}

		err := json.NewDecoder(response.Body).Decode(&bodyData)
		g.Expect(err).NotTo(HaveOccurred(),"error while decoding body: %v ", response.Body)

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
		redirect(response, request)

		var bodyData struct {
			Json map[string]string `json:"json"`
		}

		err := json.NewDecoder(response.Body).Decode(&bodyData)
		g.Expect(err).NotTo(HaveOccurred(),"error while decoding body: %v ", response.Body)

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
		redirect(response, request)

		var bodyData struct {
			Args map[string]string `json:"args"`
		}

		err := json.NewDecoder(response.Body).Decode(&bodyData)
		g.Expect(err).NotTo(HaveOccurred(),"error while decoding body: %v ", response.Body)

		g.Expect(bodyData.Args["plan_id"]).To(Equal(expectedPlan))
	})
}

func TestBadGateway(t *testing.T) {
	g := NewGomegaWithT(t)
	config.ForwardURL = "doesntexist.org"

	body := []byte{'{', '}'}
	request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/get", bytes.NewReader(body))
	request.Header = make(http.Header)
	request.Header["accept"] = []string{"application/json"}

	response := httptest.NewRecorder()
	redirect(response, request)
	got := response.Code

	want := 502
	g.Expect(got).To(Equal(want))
}
