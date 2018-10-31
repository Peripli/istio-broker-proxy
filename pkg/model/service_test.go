package model

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
)

func TestServiceInvalidJson(t *testing.T) {
	g := NewGomegaWithT(t)
	var s Service
	err := json.Unmarshal([]byte(`[]`), &s)
	g.Expect(err).To(HaveOccurred())
}

func TestServiceInvalidServiceName(t *testing.T) {
	g := NewGomegaWithT(t)
	var s Service
	err := json.Unmarshal([]byte(`{"id" : {}}`), &s)
	g.Expect(err).To(HaveOccurred())
}
