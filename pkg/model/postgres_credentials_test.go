package model

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
)

func TestCredentialIsChanged(t *testing.T) {
	g := NewGomegaWithT(t)

	credentials := PostgresCredentials{
		Hostname: "a", Port: 1, URI: "postgres://user:passwd@a:1/dbname",
	}
	credentials.Adapt([]EndpointMapping{
		EndpointMapping{
			Source: Endpoint{Host: "a", Port: 1},
			Target: Endpoint{Host: "b", Port: 2},
		}})
	g.Expect(credentials.Hostname).To(Equal("b"))
	g.Expect(credentials.Port).To(Equal(2))
	g.Expect(credentials.URI).To(Equal("postgres://user:passwd@b:2/dbname"))
}

func TestPostgresMarshalUnmarshal(t *testing.T) {
	g := NewGomegaWithT(t)

	postgresExpected := PostgresCredentials{
		Credentials: Credentials{
			AdditionalProperties: make(map[string]json.RawMessage),
			Endpoints:            []Endpoint{{Host: "a", Port: 1}},
		},
		Hostname: "a", Port: 1, URI: "postgres://user:passwd@a:1/dbname",
		WriteURL: "postgres://write", ReadURL: "postgres://read",
	}
	data, err := json.Marshal(postgresExpected.ToCredentials())
	g.Expect(err).NotTo(HaveOccurred())

	var credentials Credentials
	err = json.Unmarshal(data, &credentials)
	g.Expect(err).NotTo(HaveOccurred())

	postgres, err := PostgresCredentialsFromCredentials(credentials)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(postgres.Hostname).To(Equal(postgresExpected.Hostname))
	g.Expect(postgres.Port).To(Equal(postgresExpected.Port))
	g.Expect(postgres.URI).To(Equal(postgresExpected.URI))
	g.Expect(postgres.WriteURL).To(Equal(postgresExpected.WriteURL))
	g.Expect(postgres.ReadURL).To(Equal(postgresExpected.ReadURL))
	g.Expect(postgres.Endpoints).To(Equal(postgresExpected.Endpoints))
	g.Expect(postgres.AdditionalProperties).To(Equal(postgresExpected.AdditionalProperties))
}

func TestCredentialIsChangedToAnotherValue(t *testing.T) {
	g := NewGomegaWithT(t)

	credentials := PostgresCredentials{
		Hostname: "a", Port: 1, URI: "postgres://user:passwd@a:1/dbname",
	}
	credentials.Adapt([]EndpointMapping{
		{
			Source: Endpoint{Host: "a", Port: 1},
			Target: Endpoint{Host: "myhost", Port: 3},
		}})

	g.Expect(credentials.Hostname).To(Equal("myhost"))
	g.Expect(credentials.Port).To(Equal(3))
	g.Expect(credentials.URI).To(Equal("postgres://user:passwd@myhost:3/dbname"))
}

// Documenting weird behaviour of our implementation: We apply endpoint-mappings in place. And that means that two mappings with
// the source of the second mapping matching the target of the first mapping will be applied both in sequence.
func TestSecondMappingChangesResultOfFirstMapping(t *testing.T) {
	g := NewGomegaWithT(t)

	credentials := PostgresCredentials{
		Hostname: "a", Port: 1, URI: "postgres://user:passwd@a:1/dbname",
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
	g.Expect(credentials.URI).To(Equal("postgres://user:passwd@c:99/dbname"))

}

func TestPortDoesntMatch(t *testing.T) {
	g := NewGomegaWithT(t)

	credentials := PostgresCredentials{
		Hostname: "a", Port: 1, URI: "postgres://user:passwd@a:1/dbname",
	}
	credentials.Adapt([]EndpointMapping{
		{
			Source: Endpoint{Host: "a", Port: 99},
			Target: Endpoint{Host: "b", Port: 2},
		}})
	g.Expect(credentials.Hostname).To(Equal("a"))
	g.Expect(credentials.Port).To(Equal(1))
	g.Expect(credentials.URI).To(Equal("postgres://user:passwd@a:1/dbname"))
}

func TestHostDoesntMatch(t *testing.T) {
	g := NewGomegaWithT(t)

	credentials := PostgresCredentials{
		Hostname: "a", Port: 1, URI: "postgres://user:passwd@a:1/dbname",
	}
	credentials.Adapt([]EndpointMapping{
		EndpointMapping{
			Source: Endpoint{Host: "c", Port: 1},
			Target: Endpoint{Host: "b", Port: 2},
		}})
	g.Expect(credentials.Hostname).To(Equal("a"))
	g.Expect(credentials.Port).To(Equal(1))
	g.Expect(credentials.URI).To(Equal("postgres://user:passwd@a:1/dbname"))
}

func TestWriteUrlIsAdapted(t *testing.T) {
	g := NewGomegaWithT(t)

	credentials := PostgresCredentials{
		WriteURL: "jdbc:postgresql://10.11.19.240,10.11.19.241/e2b91324e12361f3eaeb35fe570efe1d?targetServerType=master",
		Hostname: "a", Port: 1,
	}
	credentials.Adapt([]EndpointMapping{
		EndpointMapping{
			Source: Endpoint{Host: "10.11.19.240", Port: 5432},
			Target: Endpoint{Host: "hosta", Port: 123},
		},
		EndpointMapping{
			Source: Endpoint{Host: "10.11.19.241", Port: 5432},
			Target: Endpoint{Host: "hostb", Port: 456},
		}})
	g.Expect(credentials.WriteURL).To(Equal("jdbc:postgresql://hosta:123,hostb:456/e2b91324e12361f3eaeb35fe570efe1d?targetServerType=master"))
}

func TestWriteUrlIsAdaptedWithGivenDefaultPort(t *testing.T) {
	g := NewGomegaWithT(t)

	credentials := PostgresCredentials{
		WriteURL: "jdbc:postgresql://10.11.19.240,10.11.19.241/e2b91324e12361f3eaeb35fe570efe1d?targetServerType=master",
		Hostname: "a", Port: 1,
	}
	credentials.Adapt([]EndpointMapping{
		EndpointMapping{
			Source: Endpoint{Host: "10.11.19.240", Port: 5432},
			Target: Endpoint{Host: "hosta", Port: 123},
		},
		EndpointMapping{
			Source: Endpoint{Host: "10.11.19.241", Port: 5432},
			Target: Endpoint{Host: "hostb", Port: 456},
		}})
	g.Expect(credentials.WriteURL).To(Equal("jdbc:postgresql://hosta:123,hostb:456/e2b91324e12361f3eaeb35fe570efe1d?targetServerType=master"))
}

func TestReadUrlIsAdapted(t *testing.T) {
	g := NewGomegaWithT(t)

	credentials := PostgresCredentials{
		ReadURL:  "jdbc:postgresql://10.11.19.240,10.11.19.241/e2b91324e12361f3eaeb35fe570efe1d?targetServerType=preferSlave\u0026loadBalanceHosts=true",
		Hostname: "a", Port: 1,
	}
	credentials.Adapt([]EndpointMapping{
		EndpointMapping{
			Source: Endpoint{Host: "10.11.19.240", Port: 5432},
			Target: Endpoint{Host: "hosta", Port: 123},
		},
		EndpointMapping{
			Source: Endpoint{Host: "10.11.19.241", Port: 5432},
			Target: Endpoint{Host: "hostb", Port: 456},
		}})
	g.Expect(credentials.ReadURL).To(Equal("jdbc:postgresql://hosta:123,hostb:456/e2b91324e12361f3eaeb35fe570efe1d?targetServerType=preferSlave\u0026loadBalanceHosts=true"))
}

func TestReadUrlIsUnchangedWhenSourceIsStringContainedWithinHost(t *testing.T) {
	g := NewGomegaWithT(t)

	credentials := PostgresCredentials{
		ReadURL:  "x:y://11.0.0.11,11.1.1.11/z",
		Hostname: "a", Port: 1,
	}
	credentials.Adapt([]EndpointMapping{
		EndpointMapping{
			Source: Endpoint{Host: "1.0.0.1", Port: 5432},
			Target: Endpoint{Host: "host", Port: 5432},
		}})
	g.Expect(credentials.ReadURL).To(Equal("x:y://11.0.0.11,11.1.1.11/z"))
}

func TestHostIsReplacedTwoTimes(t *testing.T) {
	g := NewGomegaWithT(t)

	credentials := PostgresCredentials{
		ReadURL:  "x:y://hosta,hostb,hosta/z",
		Hostname: "a", Port: 1,
	}
	credentials.Adapt([]EndpointMapping{
		EndpointMapping{
			Source: Endpoint{Host: "hosta", Port: 5432},
			Target: Endpoint{Host: "hostX", Port: 5432},
		}})
	g.Expect(credentials.ReadURL).To(Equal("x:y://hostX:5432,hostb,hostX:5432/z"))
}

func TestCredentialsInvalidUri(t *testing.T) {
	g := NewGomegaWithT(t)
	var credentials Credentials
	err := json.Unmarshal([]byte(`{ "uri" : 1234}`), &credentials)
	g.Expect(err).NotTo(HaveOccurred())
	_, err = PostgresCredentialsFromCredentials(credentials)
	g.Expect(err).To(HaveOccurred())
}

func TestCredentialsInvalidPort(t *testing.T) {
	g := NewGomegaWithT(t)
	var credentials Credentials
	err := json.Unmarshal([]byte(`{ "port" : [], "uri" : "postgres://"}`), &credentials)
	g.Expect(err).NotTo(HaveOccurred())
	_, err = PostgresCredentialsFromCredentials(credentials)
	g.Expect(err).To(HaveOccurred())
}

func TestCredentialsInvalidHostname(t *testing.T) {
	g := NewGomegaWithT(t)
	var credentials Credentials
	err := json.Unmarshal([]byte(`{ "hostname" : 1234, "uri" : "postgres://"}`), &credentials)
	g.Expect(err).NotTo(HaveOccurred())
	_, err = PostgresCredentialsFromCredentials(credentials)
	g.Expect(err).To(HaveOccurred())
}

func TestCredentialsInvalidReadUrl(t *testing.T) {
	g := NewGomegaWithT(t)
	var credentials Credentials
	err := json.Unmarshal([]byte(`{ "read_url" : 1234, "uri" : "postgres://"}`), &credentials)
	g.Expect(err).NotTo(HaveOccurred())
	_, err = PostgresCredentialsFromCredentials(credentials)
	g.Expect(err).To(HaveOccurred())
}

func TestCredentialsInvalidWriteUrl(t *testing.T) {
	g := NewGomegaWithT(t)
	var credentials Credentials
	err := json.Unmarshal([]byte(`{ "write_url" : 1234, "uri" : "postgres://"}`), &credentials)
	g.Expect(err).NotTo(HaveOccurred())
	_, err = PostgresCredentialsFromCredentials(credentials)
	g.Expect(err).To(HaveOccurred())
}

func TestPostgresCredentialsFromMySql(t *testing.T) {
	g := NewGomegaWithT(t)
	var credentials Credentials
	err := json.Unmarshal([]byte(`{ "hostname" : 1234, "uri" : "mysql://"}`), &credentials)
	g.Expect(err).NotTo(HaveOccurred())
	c, err := PostgresCredentialsFromCredentials(credentials)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(c).To(BeNil())
}

func TestPostgresCredentialsInvalid(t *testing.T) {
	g := NewGomegaWithT(t)
	var credentials Credentials
	err := json.Unmarshal([]byte(`{ "hostname" : "", "uri" : "postgres://"}`), &credentials)
	g.Expect(err).NotTo(HaveOccurred())
	_, err = PostgresCredentialsFromCredentials(credentials)
	g.Expect(err).To(HaveOccurred())
}
