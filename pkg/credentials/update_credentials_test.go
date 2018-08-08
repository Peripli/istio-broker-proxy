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

func TestHostAndPortDoNotMatchCredentials(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("a", "1").setEndpointMapping("nota", "2", "nota", "3")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialsAs(actual()))
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

func TestHostIsChangedWithSecondMapping(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("a", "1").setEndpointMapping("xxx", "3", "unused", "2").
		addEndpointMapping("a", "1", "yourhost", "2")
	expectedBuilder.setHostPort("yourhost", "")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "hostname"))
}

func TestAnotherHostnameThanAIsChanged(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("hugo", "4").setEndpointMapping("hugo", "4", "emil", "2")
	expectedBuilder.setHostPort("emil", "")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "hostname"))
}

func TestSecondMappingWouldChangeResultOfFirstMapping(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("a", "5").setEndpointMapping("a", "5", "b", "2").
		addEndpointMapping("b", "2", "c", "99")
	expectedBuilder.setHostPort("b", "")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "hostname"))
}

func TestPortIsChanged(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("a", "1").setEndpointMapping("a", "1", "b", "2")
	expectedBuilder.setHostPort("b", "2")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "port"))
}

func TestPortIsUnChanged(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("a", "1").setEndpointMapping("a", "2", "a", "4")
	expectedBuilder.setHostPort("a", "1")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "port"))
}

func TestUriIsChanged(t *testing.T) {
	g := NewGomegaWithT(t)

	actualBuilder.setHostPort("a", "1").setUri("postgres://user:passwd@a:1/dbname").
		setEndpointMapping("a", "1", "b", "2")
	expectedBuilder.setHostPort("b", "2").setUri("postgres://user:passwd@b:2/dbname")

	g.Expect(translateCredentials(actual())).To(haveTheSameCredentialFieldAs(expected(), "uri"))
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