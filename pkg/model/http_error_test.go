package model

import (
	"fmt"
	. "github.com/onsi/gomega"
	"net/http"
	"testing"
)

func TestHttpErrorFromError(t *testing.T) {
	g := NewGomegaWithT(t)

	err := HttpErrorFromError(fmt.Errorf("Hello Test"), http.StatusInternalServerError)
	g.Expect(err.ErrorMsg).To(Equal("Hello Test"))
	g.Expect(err.Description).To(Equal(""))
	g.Expect(err.Error()).To(Equal("Hello Test"))
	g.Expect(err.StatusCode).To(Equal(http.StatusInternalServerError))
}

func TestHttpErrorFromHttpError(t *testing.T) {
	g := NewGomegaWithT(t)

	httpError := HttpErrorFromError(HttpError{ErrorMsg: "Hello istio", Description: "xxx", StatusCode: http.StatusBadRequest}, http.StatusBadGateway)
	err := HttpErrorFromError(httpError, http.StatusBadGateway)
	g.Expect(err).To(Equal(httpError))
}

func TestHttpErrorFromResponseOK(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(HttpErrorFromResponse(200, []byte(""))).NotTo(HaveOccurred())
}

func TestHttpErrorFromResponseNotOKInvalidBody(t *testing.T) {
	g := NewGomegaWithT(t)

	err := HttpErrorFromResponse(401, []byte("Invalid body"))
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.(*HttpError).StatusCode).To(Equal(401))
	g.Expect(err.(*HttpError).ErrorMsg).To(Equal("Invalid body"))
}

func TestHttpErrorFromResponseNotOK(t *testing.T) {
	g := NewGomegaWithT(t)

	err := HttpErrorFromResponse(504, []byte(`{ "error": "my-error", "description": "my-description"}`))
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.(*HttpError).StatusCode).To(Equal(504))
	g.Expect(err.(*HttpError).ErrorMsg).To(Equal("my-error"))
	g.Expect(err.(*HttpError).Description).To(Equal("my-description"))
}