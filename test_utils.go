package main

import (
	"io"
	"net/http"
	"net/http/httptest"
)

type handlerStub struct {
	code         int
	responseBody []byte
}

func (stub handlerStub) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(stub.code)
	writer.Write(stub.responseBody)
}

func NewHandlerStub(code int, responseBody []byte) http.Handler {
	stub := handlerStub{code, responseBody}
	return stub
}

func injectClientStub(handler http.Handler) *httptest.Server {
	ts := httptest.NewServer(handler)
	client := ts.Client()
	config.httpClientFactory = func(tr *http.Transport) *http.Client {
		return client
	}
	config.httpRequestFactory = func(method string, url string, body io.Reader) (*http.Request, error) {
		return http.NewRequest(method, ts.URL, body)
	}
	return ts
}
