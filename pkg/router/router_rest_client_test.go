package router

import (
	"github.com/Peripli/istio-broker-proxy/pkg/model"
	. "github.com/onsi/gomega"
	"net/http"
	"net/url"
	"testing"
)

type TestStruct struct {
	Member1 string `json:"member1"`
	Member2 int    `json:"member2"`
}

func TestRouterRestClientPut(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := newHandlerStubWithFunc(http.StatusOK, func(body []byte) []byte {
		return []byte(`{"member1": "string","member2": 1}`)
	})
	server, routerConfig := injectClientStub(handlerStub)
	defer server.Close()
	client := &restClient{routerConfig.HTTPClientFactory(&http.Transport{}), &http.Request{URL: &url.URL{}}, *routerConfig}
	testStruct := TestStruct{"s", 10}
	err := client.Put(&testStruct).Do().Into(&testStruct)

	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(testStruct.Member1).To(Equal("string"))
	g.Expect(testStruct.Member2).To(Equal(1))
	g.Expect(handlerStub.spy.method).To(Equal(http.MethodPut))
	g.Expect(handlerStub.spy.body[0]).To(MatchJSON(`{"member1": "s","member2": 10}`))

}

func TestRouterRestClientWithBadRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := newHandlerStub(http.StatusBadRequest, []byte(`{"error" : "myerror", "description" : "mydescription"}`))
	server, routerConfig := injectClientStub(handlerStub)
	defer server.Close()
	client := &restClient{routerConfig.HTTPClientFactory(&http.Transport{}), &http.Request{URL: &url.URL{}}, *routerConfig}
	err := client.Get().Do().Error()

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.(*model.HTTPError).StatusCode).To(Equal(http.StatusBadRequest))
	g.Expect(err.(*model.HTTPError).ErrorMsg).To(Equal("myerror"))
	g.Expect(err.(*model.HTTPError).Description).To(Equal("mydescription: from call to GET http://xxxxx.xx"))
	g.Expect(handlerStub.spy.method).To(Equal(http.MethodGet))
}

func TestRouterRestClientPostWithInvalidJson(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := newHandlerStub(http.StatusOK, []byte(""))
	server, routerConfig := injectClientStub(handlerStub)
	defer server.Close()
	client := &restClient{routerConfig.HTTPClientFactory(&http.Transport{}), &http.Request{URL: &url.URL{}}, *routerConfig}
	result := TestStruct{}
	err := client.Post(&result).Do().Into(&result)

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("Can't unmarshal response from"))
	g.Expect(handlerStub.spy.method).To(Equal(http.MethodPost))
}

func TestRouterRestClientDelete(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := newHandlerStubWithFunc(http.StatusOK, func(body []byte) []byte {
		return []byte(`{"member1": "string","member2": 1}`)
	})
	server, routerConfig := injectClientStub(handlerStub)
	defer server.Close()
	client := &restClient{routerConfig.HTTPClientFactory(&http.Transport{}), &http.Request{URL: &url.URL{}}, *routerConfig}
	err := client.Delete().Do().Error()

	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(handlerStub.spy.method).To(Equal(http.MethodDelete))

}

func TestRouterRestClientWithMultilineErrorResponse(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := newHandlerStub(http.StatusConflict, []byte(`{ "error": "error: Internal server error\ndescription: \n"}`))
	server, routerConfig := injectClientStub(handlerStub)
	defer server.Close()
	client := &restClient{routerConfig.HTTPClientFactory(&http.Transport{}), &http.Request{URL: &url.URL{}}, *routerConfig}
	err := client.Get().Do().Error()

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.(*model.HTTPError).StatusCode).To(Equal(http.StatusConflict))
	g.Expect(err.(*model.HTTPError).ErrorMsg).To(Equal("error: Internal server error\ndescription: \n"))
	g.Expect(err.(*model.HTTPError).Description).To(Equal(": from call to GET http://xxxxx.xx"))
	g.Expect(handlerStub.spy.method).To(Equal(http.MethodGet))
}

func TestRouterRestClientWithInvalidJsonErrorResponse(t *testing.T) {
	g := NewGomegaWithT(t)

	handlerStub := newHandlerStub(http.StatusConflict, []byte("error: Internal server error\ndescription: \n"))
	server, routerConfig := injectClientStub(handlerStub)
	defer server.Close()
	client := &restClient{routerConfig.HTTPClientFactory(&http.Transport{}), &http.Request{URL: &url.URL{}}, *routerConfig}
	err := client.Get().Do().Error()

	g.Expect(err).To(HaveOccurred())
	g.Expect(err.(*model.HTTPError).StatusCode).To(Equal(http.StatusConflict))
	g.Expect(err.(*model.HTTPError).ErrorMsg).To(Equal("InvalidJSON"))
	g.Expect(err.(*model.HTTPError).Description).To(Equal("invalid JSON 'error: Internal server error\ndescription: \n': from call to GET http://xxxxx.xx"))
	g.Expect(handlerStub.spy.method).To(Equal(http.MethodGet))
}
