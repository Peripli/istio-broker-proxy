package model

import (
	"fmt"
	. "github.com/onsi/gomega"
	"net/http"
	"testing"
)

func TestHTTPErrorFromError(t *testing.T) {
	g := NewGomegaWithT(t)

	err := HTTPErrorFromError(fmt.Errorf("Hello Test"), http.StatusInternalServerError)
	g.Expect(err.ErrorMsg).To(Equal("InternalServerError"))
	g.Expect(err.Description).To(Equal("Hello Test"))
	g.Expect(err.Error()).To(Equal("error: 'InternalServerError', description: 'Hello Test'"))
	g.Expect(err.StatusCode).To(Equal(http.StatusInternalServerError))

	err = HTTPErrorFromError(fmt.Errorf("Hello Test"), http.StatusBadGateway)
	g.Expect(err.ErrorMsg).To(Equal("BadGateway"))
	g.Expect(err.Description).To(Equal("Hello Test"))
	g.Expect(err.Error()).To(Equal("error: 'BadGateway', description: 'Hello Test'"))
	g.Expect(err.StatusCode).To(Equal(http.StatusBadGateway))
}

func TestHTTPErrorFromHTTPError(t *testing.T) {
	g := NewGomegaWithT(t)

	httpError := HTTPErrorFromError(HTTPError{ErrorMsg: "Hello istio", Description: "xxx", StatusCode: http.StatusBadRequest}, http.StatusBadGateway)
	err := HTTPErrorFromError(httpError, http.StatusInternalServerError)
	g.Expect(err).To(Equal(httpError))
	g.Expect(err.StatusCode).To(Equal(http.StatusBadRequest))
}

func TestHTTPErrorFromResponseOK(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(HTTPErrorFromResponse(200, []byte(""), "", "", "application/json")).NotTo(HaveOccurred())
}

func TestHTTPErrorFromResponseForwardsUpstreamErrorCorrectly(t *testing.T) {
	g := NewGomegaWithT(t)

	err := HTTPErrorFromResponse(503, []byte("upstream connect error or disconnect/reset before headers"), "http://egress", "POST", "text/plain")
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.(*HTTPError).StatusCode).To(Equal(503))
	g.Expect(err.(*HTTPError).ErrorMsg).To(Equal("POST to http://egress failed"))
	g.Expect(err.(*HTTPError).Description).To(Equal("upstream connect error or disconnect/reset before headers"))
}

func TestHTTPErrorFromResponseNotOKInvalidBody(t *testing.T) {
	g := NewGomegaWithT(t)

	err := HTTPErrorFromResponse(401, []byte("Invalid body"), "http://localhost", "GET", "application/json")
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.(*HTTPError).StatusCode).To(Equal(401))
	g.Expect(err.(*HTTPError).ErrorMsg).To(Equal("InvalidJSON"))
	g.Expect(err.(*HTTPError).Description).To(Equal("invalid JSON 'Invalid body': from call to GET http://localhost"))
}

func TestHTTPErrorFromResponseNotOK(t *testing.T) {
	g := NewGomegaWithT(t)

	err := HTTPErrorFromResponse(504, []byte(`{ "error": "my-error", "description": "my-description"}`), "http://localhost", "GET", "application/json")
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.(*HTTPError).StatusCode).To(Equal(504))
	g.Expect(err.(*HTTPError).ErrorMsg).To(Equal("my-error"))
	g.Expect(err.(*HTTPError).Description).To(Equal("my-description: from call to GET http://localhost"))
}
