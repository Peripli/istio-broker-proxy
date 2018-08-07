package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestParseBody(t *testing.T) {

	dummyBody := []byte{1, 2, 3}
	request, _ := http.NewRequest(http.MethodGet, "/notused", nil)
	got := translateBody(request, dummyBody)
	want := dummyBody

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got '%v' want '%v'", got, want)
	}
}

func TestInitialInfo(t *testing.T) {
	request, _ := http.NewRequest(http.MethodGet, "/notused", nil)
	response := httptest.NewRecorder()
	info(response, request)
	got := response.Body.String()
	want := ""

	if got != want {
		t.Errorf("got '%s', want '%s'", got, want)
	}
}

func TestInfoAfterRequest(t *testing.T) {
	config.ForwardURL = "doesntexist.org"
	body := []byte{'{', '}'}
	request, _ := http.NewRequest(http.MethodGet, "/anything", bytes.NewReader(body))
	response := httptest.NewRecorder()
	redirect(response, request)

	request2, _ := http.NewRequest(http.MethodGet, "/notused", nil)
	response2 := httptest.NewRecorder()
	info(response2, request2)
	length := len(response2.Body.String())

	if length == 0 {
		t.Errorf("got no info")
	}
}

func TestDefaultPortUsed(t *testing.T) {
	expectedPort := DefaultPort

	readPort()

	if config.port != expectedPort {
		t.Errorf("wrong port: %s", config.port)
	}
}

func TestCustomPortUsed(t *testing.T) {
	oldPort := os.Getenv("PORT")
	defer func() {
		os.Setenv("PORT", oldPort)
	}()

	expectedPort := "1234"
	os.Setenv("PORT", expectedPort)

	readPort()

	if config.port != expectedPort {
		t.Errorf("wrong port: %s", config.port)
	}
}

func TestCreateNewURL(t *testing.T) {

	const internalHost = "internal-name.test"
	const externalURL = "external-name.test/cf"
	const path = "hello"

	t.Run("Test rewrite host", func(t *testing.T) {
		body := []byte{'{', '}'}
		request, _ := http.NewRequest(http.MethodGet, "https://"+internalHost+"/"+path, bytes.NewReader(body))
		request.Header = make(http.Header)
		request.Header["accept"] = []string{"application/json"}

		got := createNewUrl(externalURL, request)
		want := "https://" + externalURL + "/" + path

		if got != want {
			t.Errorf("got '%s', want '%s'", got, want)
		}
	})

	t.Run("Test rewrite host with parameter", func(t *testing.T) {
		body := []byte{'{', '}'}
		request, _ := http.NewRequest(http.MethodGet, "https://"+internalHost+"/"+path+"?debug=true", bytes.NewReader(body))
		request.Header = make(http.Header)
		request.Header["accept"] = []string{"application/json"}

		got := createNewUrl(externalURL, request)
		want := "https://" + externalURL + "/" + path + "?debug=true"

		if got != want {
			t.Errorf("got '%s', want '%s'", got, want)
		}
	})
}

func TestRedirect(t *testing.T) {

	config.ForwardURL = "httpbin.org"

	t.Run("Check return code of redirected get", func(t *testing.T) {

		body := []byte{'{', '}'}
		request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/get", bytes.NewReader(body))
		request.Header = make(http.Header)
		request.Header["accept"] = []string{"application/json"}

		response := httptest.NewRecorder()
		redirect(response, request)
		want := 200
		got := response.Code

		if got != want {
			t.Errorf("got '%d', want '%d'", got, want)
		}
	})

	t.Run("Check return code of redirected get with error code", func(t *testing.T) {

		body := []byte{'{', '}'}
		request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/status/503", bytes.NewReader(body))
		request.Header = make(http.Header)
		request.Header["accept"] = []string{"application/json"}

		response := httptest.NewRecorder()
		redirect(response, request)
		want := 503
		got := response.Code

		if got != want {
			t.Errorf("got '%d', want '%d'", got, want)
		}
	})

	t.Run("Check URL in response", func(t *testing.T) {
		body := []byte{'{', '}'}
		request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/get", bytes.NewReader(body))
		request.Header = make(http.Header)
		request.Header["accept"] = []string{"application/json"}
		//request.
		response := httptest.NewRecorder()
		redirect(response, request)

		want := "https://" + config.ForwardURL + "/get"

		var bodyData struct {
			URL string `json:"url"`
		}

		err := json.NewDecoder(response.Body).Decode(&bodyData)
		if err != nil {
			panic(err)
		}

		got := bodyData.URL

		if got != want {
			t.Errorf("got '%s', want '%s'", got, want)
		}
	})

	t.Run("Check that headers are forwarded", func(t *testing.T) {

		const testHeaderKey = "X-Broker-Api-Version"
		const testHeaderValue = "2.13"

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
		if err != nil {
			fmt.Printf("%v", response.Body)
			panic(err)
		}

		want := request.Header.Get(testHeaderKey)
		got := bodyData.Headers[testHeaderKey]
		if got != want {
			t.Errorf("got '%s', want '%s'", got, want)
		}
	})

	t.Run("Check that the request body is forwarded for PUT", func(t *testing.T) {

		body := []byte(`{"service_id":"6db542eb-8187-4afc-8a85-e08b4a3cc24e","plan_id":"c3320e0f-5866-4f14-895e-48bc92a4245c"}`)
		request, _ := http.NewRequest(http.MethodPut, "https://blahblubs.org/put", bytes.NewReader(body))
		request.Header = make(http.Header)
		request.Header.Set("accept", "application/json")
		request.Header.Set("'Content-Type", "application/json")

		response := httptest.NewRecorder()
		redirect(response, request)

		var bodyData struct {
			JSON map[string]string `json:"json"`
		}

		err := json.NewDecoder(response.Body).Decode(&bodyData)
		if err != nil {
			fmt.Printf("%v", response.Body)
			panic(err)
		}

		want := "6db542eb-8187-4afc-8a85-e08b4a3cc24e"

		got := bodyData.JSON["service_id"]
		if got != want {
			t.Errorf("got '%s', want '%s'", got, want)
		}

	})

	t.Run("Check that the request param is forwarded for DELETE", func(t *testing.T) {

		body := []byte(`{}`)
		expectedPlan := "myplan"
		request, _ := http.NewRequest(http.MethodDelete, "https://blahblubs.org/delete?plan_id="+expectedPlan, bytes.NewReader(body))
		request.Header = make(http.Header)
		request.Header.Set("accept", "application/json")
		request.Header.Set("'Content-Type", "application/json")

		response := httptest.NewRecorder()
		redirect(response, request)

		var bodyData struct {
			ARGS map[string]string `json:"args"`
		}

		err := json.NewDecoder(response.Body).Decode(&bodyData)
		if err != nil {
			t.Errorf("error while parsing json")
		}

		if bodyData.ARGS["plan_id"] != expectedPlan {
			t.Errorf("expected %s, actual %s", expectedPlan, bodyData.ARGS["plan_id"])
		}

	})

}

func TestBadGateway(t *testing.T) {

	config.ForwardURL = "doesntexist.org"

	body := []byte{'{', '}'}
	request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/get", bytes.NewReader(body))
	request.Header = make(http.Header)
	request.Header["accept"] = []string{"application/json"}

	response := httptest.NewRecorder()
	redirect(response, request)
	want := 502
	got := response.Code

	if got != want {
		t.Errorf("got '%d', want '%d'", got, want)
	}
}
