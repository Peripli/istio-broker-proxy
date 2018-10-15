package router

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

type requestSpy struct {
	method string
	url    string
	body   []string
}

type handlerStub struct {
	code    int
	handler func([]byte) []byte
	spy     requestSpy
}

func (stub handlerStub) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	bodyAsBytes, err := ioutil.ReadAll(request.Body)
	if nil != err {
		panic(err)
	}
	writer.WriteHeader(stub.code)
	writer.Write(stub.handler(bodyAsBytes))
}

func NewHandlerStub(code int, responseBody []byte) *handlerStub {
	stub := handlerStub{code, func([]byte) []byte { return responseBody }, requestSpy{}}
	return &stub
}

func NewHandlerStubWithFunc(code int, handler func([]byte) []byte) *handlerStub {
	stub := handlerStub{code, handler, requestSpy{}}
	return &stub
}

func injectClientStub(handler *handlerStub) (*httptest.Server, *RouterConfig) {

	routerConfig := RouterConfig{ForwardURL: "http://xxxxx.xx"}

	ts := httptest.NewServer(handler)
	client := ts.Client()
	routerConfig.HttpClientFactory = func(tr *http.Transport) *http.Client {
		return client
	}
	routerConfig.HttpRequestFactory = func(method string, url string, body io.Reader) (*http.Request, error) {
		handler.spy.method = method
		handler.spy.url = url
		buf := new(bytes.Buffer)
		buf.ReadFrom(body)

		handler.spy.body = append(handler.spy.body, buf.String()) // Does a complete copy of the bytes in the buffer.
		return http.NewRequest(method, ts.URL, bytes.NewReader(buf.Bytes()))
	}
	return ts, &routerConfig
}
