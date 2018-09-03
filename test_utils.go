package main

import (
	"io"
	"net/http"
	"net/http/httptest"
)

type requestSpy struct {
	method string
	url    string
}

type handlerStub struct {
	code         int
	responseBody []byte
	spy          requestSpy
}

func (stub handlerStub) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(stub.code)
	writer.Write(stub.responseBody)
}

func NewHandlerStub(code int, responseBody []byte) *handlerStub {
	stub := handlerStub{code, responseBody, requestSpy{}}
	return &stub
}

func injectClientStub(handler *handlerStub) *httptest.Server {
	ts := httptest.NewServer(handler)
	client := ts.Client()
	config.httpClientFactory = func(tr *http.Transport) *http.Client {
		return client
	}
	config.httpRequestFactory = func(method string, url string, body io.Reader) (*http.Request, error) {
		handler.spy.method = method
		handler.spy.url = url
		return http.NewRequest(method, ts.URL, body)
	}
	return ts
}
