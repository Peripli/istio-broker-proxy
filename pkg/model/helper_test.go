package model

import (
	"encoding/json"
	. "github.com/onsi/gomega"
	"testing"
)

func TestRemoveIntOrStringProperty(t *testing.T) {
	g := NewGomegaWithT(t)
	body := []byte(`{
                "port1": "5432",
                "port2": 1234
            }`)

	var additionalProperties map[string]json.RawMessage
	err := json.Unmarshal(body, &additionalProperties)
	g.Expect(err).NotTo(HaveOccurred())
	var port int
	err = removeIntOrStringProperty(additionalProperties, "port1", &port)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(port).To(Equal(5432))

	err = removeIntOrStringProperty(additionalProperties, "port2", &port)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(port).To(Equal(1234))

	port = 0
	err = removeIntOrStringProperty(additionalProperties, "port3", &port)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(port).To(Equal(0))
}
