// +build integration

package integration

import (
	"bytes"
	"crypto/tls"
	. "github.com/onsi/gomega"
	"net/http"
	"testing"
)

const baseUrl = "https://istio-broker.cfapps.dev01.aws.istio.sapcloud.io"

func TestAdaptCredentialsWithInvalidRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	body := []byte(`
                  {
                  "credentials": {
                   "dbname": "yLO2WoE0-mCcEppn",
                   "hostname": "10.11.241.0",
                   "password": "redacted",
                   "port": "47637",
                   "ports": {
                    "5432/tcp": "47637"
                   },
                   "uri": "postgres://mma4G8N0isoxe17v:redacted@10.11.241.0:47637/yLO2WoE0-mCcEppn",
                   "username": "mma4G8N0isoxe17v"
                  }
                  }`)
	client := http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	request, err := http.NewRequest(http.MethodPut, baseUrl+"/v2/service_instances/1/service_bindings/2/adapt_credentials", bytes.NewReader(body))
	g.Expect(err).NotTo(HaveOccurred())

	response, err := client.Do(request)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(response.StatusCode).To(Equal(400))

}

func TestAdaptCredentialsWithValidRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	body := []byte(`
                  {
                  "credentials": {
                   "dbname": "yLO2WoE0-mCcEppn",
                   "hostname": "10.11.241.0",
                   "password": "redacted",
                   "port": "47637",
                   "ports": {
                    "5432/tcp": "47637"
                   },
                   "uri": "postgres://mma4G8N0isoxe17v:redacted@10.11.241.0:47637/yLO2WoE0-mCcEppn",
                   "username": "mma4G8N0isoxe17v"
                  },
                  "endpoint_mappings": [{
                    "source": {"host": "10.11.241.0", "port": 47637},
                    "target": {"host": "appnethost", "port": 9876}
                  	}]
                  }`)
	client := http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	request, err := http.NewRequest(http.MethodPut, baseUrl+"/v2/service_instances/1/service_bindings/2/adapt_credentials", bytes.NewReader(body))
	g.Expect(err).NotTo(HaveOccurred())

	response, err := client.Do(request)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(response.StatusCode).To(Equal(200))

}
