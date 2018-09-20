package model

import (
	"encoding/json"
	. "github.com/onsi/gomega"
	"testing"
)

func TestEndpointUnmarshalPortAsString(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{
                "host": "10.11.19.245",
                "network_id": "SF",
                "port": "5432"
            }`)
	var ep Endpoint
	err := json.Unmarshal(body, &ep)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ep.Port).To(Equal(5432))
}

func TestEndpointUnmarshalPortAsIntg(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{
                "host": "10.11.19.245",
                "network_id": "SF",
                "port": 5432
            }`)
	var ep Endpoint
	err := json.Unmarshal(body, &ep)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ep.Port).To(Equal(5432))
}

func TestEndpointHostIsIP(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{
                "host": "my-host",
                "network_id": "SF",
                "port": "5432"
            }`)
	var ep Endpoint
	err := json.Unmarshal(body, &ep)
	g.Expect(err).ToNot(HaveOccurred())
}
