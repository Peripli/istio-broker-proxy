package credentials

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestCredentialMatcher(t *testing.T) {
	g := NewGomegaWithT(t)
	g.Expect(exampleRequest).To(haveTheSameCredentialsAs(exampleRequest))
	g.Expect(exampleRequest).To(haveTheSameCredentialFieldAs(exampleRequest, "host"))
	g.Expect(exampleRequest).NotTo(haveTheSameCredentialsAs(ExampleRequestHaPostgres))
}

func TestCredentialMatcherMessage(t *testing.T) {
	g := NewGomegaWithT(t)

	matcher := CredentialMatcher{expected: "a", checkField: "b"}
	g.Expect(matcher.FailureMessage("c")).To(MatchRegexp("^Credentials do not match"))
	g.Expect(matcher.NegatedFailureMessage("c")).To(MatchRegexp("^Credentials are unchanged"))
}

func TestCredentialMatcherErrorHandling(t *testing.T) {
	g := NewGomegaWithT(t)

	matcher := CredentialMatcher{expected: "a", checkField: "b"}
	_, err := matcher.Match("x")
	g.Expect(err).Should(HaveOccurred())
}

func TestCredentialMatcherErrorHandlingInExpectedValue(t *testing.T) {
	g := NewGomegaWithT(t)

	matcher := CredentialMatcher{expected: "a", checkField: "b"}
	_, err := matcher.Match(`{ "credentials": {} }`)
	g.Expect(err).Should(HaveOccurred())
}
