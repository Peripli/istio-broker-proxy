package model

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
)

func TestPlan(t *testing.T) {
	g := NewGomegaWithT(t)
	var plan Plan
	examplePlan := `{ "name": "fake-plan-1", "metadata": { "max_storage_tb": 5} }`
	err := json.Unmarshal([]byte(examplePlan), &plan)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(plan.MetaData).NotTo(BeEmpty())
	marshaledPlan, err := json.Marshal(&plan)
	g.Expect(marshaledPlan).To(MatchJSON(examplePlan))
}
