package router

import (
	"errors"
	"github.com/Peripli/istio-broker-proxy/pkg/model"
	. "github.com/onsi/gomega"
	"testing"
)

type TestInterceptor struct {
	noOpInterceptor
}

func (c TestInterceptor) PreBind(request model.BindRequest) (*model.BindRequest, error) {
	return nil, errors.New("error")
}

func TestInterceptedOsbClient(t *testing.T) {
	g := NewGomegaWithT(t)

	interceptedOsbClient := InterceptedOsbClient{Interceptor: TestInterceptor{}}
	_, err := interceptedOsbClient.Bind("test", &model.BindRequest{})
	g.Expect(err).To(HaveOccurred())
}
