package main

import (
	"testing"
	"reflect"
	"net/http"
	"net/http/httptest"
	"bytes"
	"encoding/json"
)

func TestParseBody(t *testing.T) {

	testArray := []byte{1,2,3}
	got := parseBody(testArray)
	want := []byte{1,2,3}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got '%v' want '%v'", got, want)
	}
}

func TestHelloWorld(t *testing.T) {
	request, _ := http.NewRequest(http.MethodGet, "/hello", nil)
	response := httptest.NewRecorder()
	helloWorld(response,request)
	got := response.Body.String()
	want := "Hello World"

	if got != want {
		t.Errorf("got '%s', want '%s'", got, want)
	}
}

func TestCreateNewURL(t *testing.T) {

	const internalHost = "internal-name.test"
	const externalURL = "external-name.test/cf"
	const path = "hello"

	t.Run("rewrite host", func(t *testing.T) {
		body := []byte{'{','}'}
		request, _ := http.NewRequest(http.MethodGet, "https://" + internalHost +"/" + path, bytes.NewReader(body) )
		request.Header = make(http.Header)
		request.Header["accept"] = []string{"application/json"}

		got := createNewUrl(externalURL, request)
		want := "https://" + externalURL + "/" + path

		if got != want {
			t.Errorf("got '%s', want '%s'", got, want)
		}
	})
}

func TestRedirectGet(t *testing.T){

	config.ForwardURL = "httpbin.org"

	body := []byte{'{','}'}
	request, _ := http.NewRequest(http.MethodGet, "https://blahblubs.org/get", bytes.NewReader(body) )
	request.Header = make(http.Header)
	request.Header["accept"] = []string{"application/json"}
	//request.
	response := httptest.NewRecorder()
	redirect(response, request)

	t.Run("Return Code",func(t *testing.T) {

		want := 200
		got := response.Code

		if got != want {
			t.Errorf("got '%d', want '%d'", got, want)
		}
	})

	t.Run("URL",func(t *testing.T) {
		want := "https://" + config.ForwardURL +"/get"

		var bodyData struct {
			// httpbin.org sends back key/value pairs, no map[string][]string
			URL  string               `json:"url"`
		}

		err := json.NewDecoder(response.Body).Decode(&bodyData)
		if err != nil {
			panic(err)
		}

		got := bodyData.URL
		//responseDump, _ := httputil.DumpResponse(response., false)
		if got != want {
			t.Errorf("got '%v', want '%s'", got, want)
		}
	})
}