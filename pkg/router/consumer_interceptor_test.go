package router

import (
	. "github.com/onsi/gomega"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/model"
	"github.infra.hana.ondemand.com/istio/istio-broker/pkg/profiles"
	"testing"
)

func TestConsumerPreBind(t *testing.T) {
	g := NewGomegaWithT(t)

	consumer := NewConsumerInterceptor(ConsumerConfig{ConsumerId: "consumer-id"})
	request := consumer.preBind(model.BindRequest{})
	g.Expect(request.NetworkData.NetworkProfileId).To(Equal(profiles.NetworkProfile))
	g.Expect(request.NetworkData.Data.ConsumerId).To(Equal("consumer-id"))

}
