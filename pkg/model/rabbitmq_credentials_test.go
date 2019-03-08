package model

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
)

func TestCredentialIsChangedRMQ(t *testing.T) {
	g := NewGomegaWithT(t)

	credentials := RabbitMQCredentials{
		Hostname: "a", Port: 1, URI: "amqp://user:passwd@a:1",
	}
	credentials.Adapt([]EndpointMapping{
		{
			Source: Endpoint{Host: "a", Port: 1},
			Target: Endpoint{Host: "b", Port: 2},
		}})
	g.Expect(credentials.Hostname).To(Equal("b"))
	g.Expect(credentials.Port).To(Equal(2))
	g.Expect(credentials.URI).To(Equal("amqp://user:passwd@b:2"))
}

func TestPostgresMarshalUnmarshalRMQ(t *testing.T) {
	g := NewGomegaWithT(t)

	rabbitmqExpected := RabbitMQCredentials{
		Credentials: Credentials{
			AdditionalProperties: make(map[string]json.RawMessage),
			Endpoints:            []Endpoint{{Host: "a", Port: 1}},
		},
		Hostname: "a", Port: 1, URI: "amqp://user:passwd@a:1",
	}
	data, err := json.Marshal(rabbitmqExpected.ToCredentials())
	g.Expect(err).NotTo(HaveOccurred())

	var credentials Credentials
	err = json.Unmarshal(data, &credentials)
	g.Expect(err).NotTo(HaveOccurred())

	rabbitmq, err := RabbitMQCredentialsFromCredentials(credentials)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(rabbitmq.Hostname).To(Equal(rabbitmqExpected.Hostname))
	g.Expect(rabbitmq.Port).To(Equal(rabbitmqExpected.Port))
	g.Expect(rabbitmq.URI).To(Equal(rabbitmqExpected.URI))

	g.Expect(rabbitmq.Endpoints).To(Equal(rabbitmqExpected.Endpoints))
	g.Expect(rabbitmq.AdditionalProperties).To(Equal(rabbitmqExpected.AdditionalProperties))
}

func TestCredentialIsChangedToAnotherValueRMQ(t *testing.T) {
	g := NewGomegaWithT(t)

	credentials := RabbitMQCredentials{
		Hostname: "a", Port: 1, URI: "amqp://user:passwd@a:1",
	}
	credentials.Adapt([]EndpointMapping{
		{
			Source: Endpoint{Host: "a", Port: 1},
			Target: Endpoint{Host: "myhost", Port: 3},
		}})

	g.Expect(credentials.Hostname).To(Equal("myhost"))
	g.Expect(credentials.Port).To(Equal(3))
	g.Expect(credentials.URI).To(Equal("amqp://user:passwd@myhost:3"))
}

// Documenting weird behaviour of our implementation: We apply endpoint-mappings in place. And that means that two mappings with
// the source of the second mapping matching the target of the first mapping will be applied both in sequence.
func TestSecondMappingChangesResultOfFirstMappingRMQ(t *testing.T) {
	g := NewGomegaWithT(t)

	credentials := RabbitMQCredentials{
		Hostname: "a", Port: 1, URI: "amqp://user:passwd@a:1",
	}
	credentials.Adapt([]EndpointMapping{
		{
			Source: Endpoint{Host: "a", Port: 1},
			Target: Endpoint{Host: "b", Port: 2},
		},
		{
			Source: Endpoint{Host: "b", Port: 2},
			Target: Endpoint{Host: "c", Port: 99},
		},
	})

	g.Expect(credentials.Hostname).To(Equal("c"))
	g.Expect(credentials.Port).To(Equal(99))
	g.Expect(credentials.URI).To(Equal("amqp://user:passwd@c:99"))

}

func TestPortDoesntMatchRMQ(t *testing.T) {
	g := NewGomegaWithT(t)

	credentials := RabbitMQCredentials{
		Hostname: "a", Port: 1, URI: "amqp://user:passwd@a:1",
	}
	credentials.Adapt([]EndpointMapping{
		{
			Source: Endpoint{Host: "a", Port: 99},
			Target: Endpoint{Host: "b", Port: 2},
		}})
	g.Expect(credentials.Hostname).To(Equal("a"))
	g.Expect(credentials.Port).To(Equal(1))
	g.Expect(credentials.URI).To(Equal("amqp://user:passwd@a:1"))
}

func TestHostDoesntMatchRMQ(t *testing.T) {
	g := NewGomegaWithT(t)

	credentials := RabbitMQCredentials{
		Hostname: "a", Port: 1, URI: "amqp://user:passwd@a:1",
	}
	credentials.Adapt([]EndpointMapping{
		{
			Source: Endpoint{Host: "c", Port: 1},
			Target: Endpoint{Host: "b", Port: 2},
		}})
	g.Expect(credentials.Hostname).To(Equal("a"))
	g.Expect(credentials.Port).To(Equal(1))
	g.Expect(credentials.URI).To(Equal("amqp://user:passwd@a:1"))
}

func TestCredentialsInvalidUriRMQ(t *testing.T) {
	g := NewGomegaWithT(t)
	var credentials Credentials
	err := json.Unmarshal([]byte(`{ "uri" : 1234}`), &credentials)
	g.Expect(err).NotTo(HaveOccurred())
	_, err = RabbitMQCredentialsFromCredentials(credentials)
	g.Expect(err).To(HaveOccurred())
}

func TestCredentialsInvalidPortRMQ(t *testing.T) {
	g := NewGomegaWithT(t)
	var credentials Credentials
	err := json.Unmarshal([]byte(`{ "port" : [], "uri" : "amqp://"}`), &credentials)
	g.Expect(err).NotTo(HaveOccurred())
	_, err = RabbitMQCredentialsFromCredentials(credentials)
	g.Expect(err).To(HaveOccurred())
}

func TestCredentialsInvalidHostnameRMQ(t *testing.T) {
	g := NewGomegaWithT(t)
	var credentials Credentials
	err := json.Unmarshal([]byte(`{ "hostname" : 1234, "uri" : "amqp://"}`), &credentials)
	g.Expect(err).NotTo(HaveOccurred())
	_, err = RabbitMQCredentialsFromCredentials(credentials)
	g.Expect(err).To(HaveOccurred())
}

func TestCredentialsInvalidReadUrlRMQ(t *testing.T) {
	g := NewGomegaWithT(t)
	var credentials Credentials
	err := json.Unmarshal([]byte(`{ "read_url" : 1234, "uri" : "amqp://"}`), &credentials)
	g.Expect(err).NotTo(HaveOccurred())
	_, err = RabbitMQCredentialsFromCredentials(credentials)
	g.Expect(err).To(HaveOccurred())
}

func TestCredentialsInvalidWriteUrlRMQ(t *testing.T) {
	g := NewGomegaWithT(t)
	var credentials Credentials
	err := json.Unmarshal([]byte(`{ "write_url" : 1234, "uri" : "amqp://"}`), &credentials)
	g.Expect(err).NotTo(HaveOccurred())
	_, err = RabbitMQCredentialsFromCredentials(credentials)
	g.Expect(err).To(HaveOccurred())
}

func TestRabbitMQCredentialsInvalidRMQ(t *testing.T) {
	g := NewGomegaWithT(t)
	var credentials Credentials
	err := json.Unmarshal([]byte(`{ "hostname" : "", "uri" : "amqp://"}`), &credentials)
	g.Expect(err).NotTo(HaveOccurred())
	_, err = RabbitMQCredentialsFromCredentials(credentials)
	g.Expect(err).To(HaveOccurred())
}

func TestRabbitMQCredentialsFromMySql(t *testing.T) {
	g := NewGomegaWithT(t)
	var credentials Credentials
	err := json.Unmarshal([]byte(`{ "hostname" : 1234, "uri" : "mysql://"}`), &credentials)
	g.Expect(err).NotTo(HaveOccurred())
	c, err := RabbitMQCredentialsFromCredentials(credentials)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(c).To(BeNil())
}
