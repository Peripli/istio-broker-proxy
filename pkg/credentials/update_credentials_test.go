package credentials

import (
	. "github.com/onsi/gomega"
	"testing"
)

var (
	actualBuilder   *credentialBuilder = newCredentialBuilder()
	expectedBuilder *credentialBuilder = newCredentialBuilder()
)

func actual() string {
	return actualBuilder.build()
}

func expected() string {
	return expectedBuilder.build()
}

func TestHostIsChanged(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("a", "1").setEndpointMapping("a", "1", "b", "2")
	expectedBuilder.setHostPort("b", "")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "hostname"))
}

func TestHostIsChangedToAnotherValue(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("a", "2").setEndpointMapping("a", "2", "myhost", "2")
	expectedBuilder.setHostPort("myhost", "")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "hostname"))
}

func TestAnotherHostnameThanAIsChanged(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("hugo", "4").setEndpointMapping("hugo", "4", "emil", "2")
	expectedBuilder.setHostPort("emil", "")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "hostname"))
}

// Documenting weird behaviour of our implementation: We apply endpoint-mappings in place. And that means that two mappings with
// the source of the second mapping matching the target of the first mapping will be applied both in sequence.
func TestSecondMappingChangesResultOfFirstMapping(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("a", "5").setEndpointMapping("a", "5", "b", "2").
		addEndpointMapping("b", "2", "c", "99")
	expectedBuilder.setHostPort("c", "")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "hostname"))
}

func TestPortIsChanged(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("a", "1").setEndpointMapping("a", "1", "b", "2")
	expectedBuilder.setHostPort("b", "2")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "port"))
}

func TestUriIsChanged(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("a", "1").setUri("postgres://user:passwd@a:1/dbname").
		setEndpointMapping("a", "1", "b", "2")
	expectedBuilder.setHostPort("b", "2").setUri("postgres://user:passwd@b:2/dbname")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "uri"))
}

func TestOnlyUriIsChanged(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("xy", "99").setUri("postgres://user:passwd@a:1/dbname").
		setEndpointMapping("a", "1", "b", "2")
	expectedBuilder.setHostPort("xy", "99").setUri("postgres://user:passwd@b:2/dbname")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "hostname"))
	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "port"))
	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "uri"))
}

func TestPortDoesntMatch(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("a", "99").setUri("postgres://user:passwd@a:1/dbname").
		setEndpointMapping("a", "1", "b", "2")
	expectedBuilder.setHostPort("a", "99").setUri("postgres://user:passwd@b:2/dbname")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "hostname"))
	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "port"))
	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "uri"))
}

func TestUriIsNotChangedIfItContainsDifferentHost(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("a", "1").setUri("postgres://user:passwd@b:2/dbname").
		setEndpointMapping("a", "1", "c", "3")
	expectedBuilder.setHostPort("c", "3").setUri("postgres://user:passwd@b:2/dbname")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "uri"))
}

func TestTwoEndpointsAreApplied(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("a", "1").setUri("postgres://user:passwd@b:2/dbname").
		setEndpointMapping("a", "1", "c", "3").
		addEndpointMapping("b", 2, "d", 4)
	expectedBuilder.setHostPort("c", "3").setUri("postgres://user:passwd@d:4/dbname")

	translatedCredentials := translateCredentials(actual())
	g.Expect(translatedCredentials).To(haveTheSameCredentialFieldAs(expected(), "uri"))
	g.Expect(translatedCredentials).To(HaveTheEndpoint("c", "3"))
	g.Expect(translatedCredentials).To(HaveTheEndpoint("d", "4"))
}

func TestUriIsChangedToTargetHostAndPort(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("a", "1").setUri("postgres://user:passwd@a:1/dbname").
		setEndpointMapping("a", "1", "myhost", 3)
	expectedBuilder.setHostPort("b", "2").setUri("postgres://user:passwd@myhost:3/dbname")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "uri"))
}

func TestUriIsChangedToTargetHostAndPortWithInteger(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("a", "1").setUri("postgres://user:passwd@a:1/dbname").
		setEndpointMapping("a", 1, "myhost", 3)
	expectedBuilder.setHostPort("b", "2").setUri("postgres://user:passwd@myhost:3/dbname")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "uri"))
}

func TestAnotherUriIsChanged(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("a", "1").setUri("postgres://me:em@a:1/dbname").
		setEndpointMapping("a", 1, "myhost", 3)
	expectedBuilder.setHostPort("b", "2").setUri("postgres://me:em@myhost:3/dbname")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "uri"))
}

func TestExampleRequestFromBacklogItem(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(translateCredentials(exampleRequest)).To(haveTheSameCredentialsAs(`{
    "credentials": {
 "dbname": "yLO2WoE0-mCcEppn",
 "hostname": "appnethost",
 "password": "redacted",
 "port": 9876,
 "ports": {
  "5432/tcp": "47637"
 },
 "uri": "postgres://mma4G8N0isoxe17v:redacted@appnethost:9876/yLO2WoE0-mCcEppn",
 "username": "mma4G8N0isoxe17v"
  }
}`))
}

func TestEndpointMappingGetsRemovedAfterApplying(t *testing.T) {
	g := NewGomegaWithT(t)

	translatedRequest := translateCredentials(exampleRequest)

	g.Expect(hasEndpointMappings(exampleRequest)).To(BeTrue())
	g.Expect(hasEndpointMappings(translatedRequest)).To(BeFalse())
}

func TestEndpointIsAddedAfterApplying(t *testing.T) {
	g := NewGomegaWithT(t)

	translatedRequest := translateCredentials(exampleRequest)

	g.Expect(exampleRequest).NotTo(HaveTheEndpoint("appnethost", "9876"))
	g.Expect(translatedRequest).To(HaveTheEndpoint("appnethost", "9876"))
}

func TestWriteUrlIsAdapted(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("a", "1").setWriteUrl("jdbc:postgresql://10.11.19.240,10.11.19.241/e2b91324e12361f3eaeb35fe570efe1d?targetServerType=master").
		setEndpointMapping("10.11.19.240", 5432, "hosta", 123).addEndpointMapping("10.11.19.241", 5432, "hostb", 456)
	expectedBuilder.setHostPort("b", "2").setWriteUrl("jdbc:postgresql://hosta:123,hostb:456/e2b91324e12361f3eaeb35fe570efe1d?targetServerType=master")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "write_url"))
}

func TestWriteUrlIsAdaptedWithGivenPort(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("a", "1").setWriteUrl("jdbc:postgresql://10.11.19.240:5433,10.11.19.241/e2b91324e12361f3eaeb35fe570efe1d?targetServerType=master").
		setEndpointMapping("10.11.19.240", 5433, "hosta", 123).addEndpointMapping("10.11.19.241", 5432, "hostb", 456)
	expectedBuilder.setHostPort("b", "2").setWriteUrl("jdbc:postgresql://hosta:123,hostb:456/e2b91324e12361f3eaeb35fe570efe1d?targetServerType=master")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "write_url"))
}

func TestWriteUrlIsAdaptedWithGivenDefaultPort(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("a", "1").setWriteUrl("jdbc:postgresql://10.11.19.240:5432,10.11.19.241/e2b91324e12361f3eaeb35fe570efe1d?targetServerType=master").
		setEndpointMapping("10.11.19.240", 5432, "hosta", 123).addEndpointMapping("10.11.19.241", 5432, "hostb", 456)
	expectedBuilder.setHostPort("b", "2").setWriteUrl("jdbc:postgresql://hosta:123,hostb:456/e2b91324e12361f3eaeb35fe570efe1d?targetServerType=master")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "write_url"))
}

func TestWriteUrlIsAdaptedWithGivenPortAsString(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("a", "1").setWriteUrl("jdbc:postgresql://10.11.19.240:5433,10.11.19.241/e2b91324e12361f3eaeb35fe570efe1d?targetServerType=master").
		setEndpointMapping("10.11.19.240", "5433", "hosta", 123).addEndpointMapping("10.11.19.241", 5432, "hostb", 456)
	expectedBuilder.setHostPort("b", "2").setWriteUrl("jdbc:postgresql://hosta:123,hostb:456/e2b91324e12361f3eaeb35fe570efe1d?targetServerType=master")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "write_url"))
}

func TestReadUrlIsAdapted(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("a", "1").setReadUrl("jdbc:postgresql://10.11.19.240,10.11.19.241/e2b91324e12361f3eaeb35fe570efe1d?targetServerType=preferSlave\u0026loadBalanceHosts=true").
		setEndpointMapping("10.11.19.240", 5432, "hosta", 123).addEndpointMapping("10.11.19.241", 5432, "hostb", 456)
	expectedBuilder.setHostPort("b", "2").setReadUrl("jdbc:postgresql://hosta:123,hostb:456/e2b91324e12361f3eaeb35fe570efe1d?targetServerType=preferSlave\u0026loadBalanceHosts=true")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "read_url"))
}
