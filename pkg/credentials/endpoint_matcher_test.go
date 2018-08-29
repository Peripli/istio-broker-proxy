package credentials

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestEndpointMatcher(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(`{"endpoints": [{"host": "myhost", "port": "1234"}] }`).To(haveTheEndpoint("myhost", "1234"))
	g.Expect(`{"endpoints": [{"host": "myhost", "port": "1234"}] }`).NotTo(haveTheEndpoint("myhost", "12345"))
}

func TestEndpointMatcherMessage(t *testing.T) {
	g := NewGomegaWithT(t)

	matcher := EndpointMatcher{"host", "1234"}
	g.Expect(matcher.FailureMessage("c")).To(MatchRegexp("^Endpoint not found"))
	g.Expect(matcher.NegatedFailureMessage("c")).To(MatchRegexp("^Endpoint found, but not expected"))
}

func TestEndpointMatcherErrorHandling(t *testing.T) {
	g := NewGomegaWithT(t)

	matcher := EndpointMatcher{"host", "1234"}
	_, err := matcher.Match("x")
	g.Expect(err).Should(HaveOccurred())
}

func TestEndpointMatcherErrorHandlingInExpectedValue(t *testing.T) {
	g := NewGomegaWithT(t)

	matcher := EndpointMatcher{"host", "1234"}
	_, err := matcher.Match(`{ "invalidjson }`)
	g.Expect(err).Should(HaveOccurred())
}

func TestEndpointMatcherEmptyJson(t *testing.T) {
	g := NewGomegaWithT(t)

	matcher := EndpointMatcher{"host", "1234"}
	match, _ := matcher.Match(`{}`)
	g.Expect(match).To(BeFalse())
}

func TestEndpointMatcherWrongType(t *testing.T) {
	g := NewGomegaWithT(t)

	matcher := EndpointMatcher{"host", "1234"}
	match, _ := matcher.Match(`{"endpoints": 0}`)
	g.Expect(match).To(BeFalse())
}

func TestHasEndpointMappingsTrue(t *testing.T) {
	g := NewGomegaWithT(t)

	ok := hasEndpointMappings(`{"endpoint_mappings": {}}`)
	g.Expect(ok).To(BeTrue())
}

func TestHasEndpointMappingsFalse(t *testing.T) {
	g := NewGomegaWithT(t)

	ok := hasEndpointMappings(`{"endpoints": {}}`)
	g.Expect(ok).To(BeFalse())
}

func TestHasEndpointMappingsError(t *testing.T) {
	g := NewGomegaWithT(t)

	ok := hasEndpointMappings(` "invalidjson }`)
	g.Expect(ok).To(BeFalse())
}
