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
	tr     *http.Transport
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
	writer.Header().Add("Content-Type","application/json")
	writer.WriteHeader(stub.code)
	writer.Write(stub.handler(bodyAsBytes))
}

func newHandlerStub(code int, responseBody []byte) *handlerStub {
	stub := handlerStub{code, func([]byte) []byte { return responseBody }, requestSpy{}}
	return &stub
}

func newHandlerStubWithFunc(code int, handler func([]byte) []byte) *handlerStub {
	stub := handlerStub{code, handler,requestSpy{}}
	return &stub
}

func injectClientStub(handler *handlerStub) (*httptest.Server, *Config) {

	routerConfig := Config{ForwardURL: "http://xxxxx.xx"}

	ts := httptest.NewServer(handler)
	client := ts.Client()
	routerConfig.HTTPClientFactory = func(tr *http.Transport) *http.Client {
		handler.spy.tr = tr
		return client
	}
	routerConfig.HTTPRequestFactory = func(method string, url string, header http.Header, body io.Reader) (*http.Request, error) {
		handler.spy.method = method
		handler.spy.url = url
		buf := new(bytes.Buffer)
		buf.ReadFrom(body)

		handler.spy.body = append(handler.spy.body, buf.String()) // Does a complete copy of the bytes in the buffer.
		return http.NewRequest(method, ts.URL, bytes.NewReader(buf.Bytes()))
	}
	return ts, &routerConfig
}
