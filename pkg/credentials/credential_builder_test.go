package credentials

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestCredentialBuilder(t *testing.T) {
	g := NewGomegaWithT(t)
	b := newCredentialBuilder()

	b.setHostPort("1.2.3.4", "5")
	b.setUri("postgres://")

	b.setEndpointMapping("a", "1", "b", "2")

	asJson := b.build()

	g.Expect(asJson).To(Equal(`{"credentials":{"hostname":"1.2.3.4","port":"5","uri":"postgres://"},"endpoint_mappings":[{"source":{"host":"a","port":"1"},"target":{"host":"b","port":"2"}}]}`))
}

func TestCredentialBuilderWithIntPort(t *testing.T) {
	g := NewGomegaWithT(t)
	b := newCredentialBuilder()
	b.setHostPort("1.2.3.4", 5)
	b.setUri("postgres://")

	b.setEndpointMapping("a", 1, "b", 2)

	asJson := b.build()

	g.Expect(asJson).To(Equal(`{"credentials":{"hostname":"1.2.3.4","port":5,"uri":"postgres://"},"endpoint_mappings":[{"source":{"host":"a","port":1},"target":{"host":"b","port":2}}]}`))
}
