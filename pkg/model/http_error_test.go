package model

import (
	"fmt"
	. "github.com/onsi/gomega"
	"net/http"
	"testing"
)

func TestHttpErrorFromError(t *testing.T) {
	g := NewGomegaWithT(t)

	err := HttpErrorFromError(fmt.Errorf("Hello Test"))
	g.Expect(err.Message).To(Equal("Hello Test"))
	g.Expect(err.Description).To(Equal(""))
	g.Expect(err.Error()).To(Equal("Hello Test"))
	g.Expect(err.Status).To(Equal(http.StatusInternalServerError))
}

func TestHttpErrorFromHttpError(t *testing.T) {
	g := NewGomegaWithT(t)

	httpError := HttpErrorFromError(HttpError{Message: "Hello istio", Description: "xxx", Status: http.StatusBadRequest})
	err := HttpErrorFromError(httpError)
	g.Expect(err).To(Equal(httpError))
}
