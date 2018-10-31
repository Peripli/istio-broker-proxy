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

func TestRemoveIntOrStringPropertyInvalidJson(t *testing.T) {
	g := NewGomegaWithT(t)

	var port int
	err := removeIntOrStringProperty(map[string]json.RawMessage{
		"port": json.RawMessage([]byte(`{`)),
	}, "port", &port)
	g.Expect(err).To(HaveOccurred())

}

type MyAdditionalProperties struct {
	AdditionalProperties AdditionalProperties
	StringMember         string
	IntMember            int
}

func TestAdditionalProperties(t *testing.T) {
	g := NewGomegaWithT(t)
	var testStruct MyAdditionalProperties

	data := `
{
  "stringmember": "string",
  "intmember" : 3,
  "additional1" : {},
  "additional2" : "test"
}
`
	err := testStruct.AdditionalProperties.UnmarshalJSON([]byte(data), map[string]interface{}{
		"stringmember": &testStruct.StringMember,
		"intmember":    &testStruct.IntMember,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(testStruct.StringMember).To(Equal("string"))
	g.Expect(testStruct.IntMember).To(Equal(3))
	g.Expect(testStruct.AdditionalProperties["additional1"]).To(Equal(json.RawMessage([]byte(`{}`))))
	g.Expect(testStruct.AdditionalProperties["additional2"]).To(Equal(json.RawMessage([]byte(`"test"`))))

	body, err := testStruct.AdditionalProperties.MarshalJSON(map[string]interface{}{
		"stringmember": &testStruct.StringMember,
		"intmember":    &testStruct.IntMember,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(body).To(MatchJSON(data))
}

func TestAdditionalPropertiesInvalid(t *testing.T) {
	g := NewGomegaWithT(t)
	var additionalProperties AdditionalProperties

	err := additionalProperties.UnmarshalJSON([]byte(`[]`), make(map[string]interface{}, 0))
	g.Expect(err).To(HaveOccurred())
}

func TestAdditionalPropertiesInvalidMember(t *testing.T) {
	g := NewGomegaWithT(t)
	var additionalProperties AdditionalProperties
	var endpoint int
	err := additionalProperties.UnmarshalJSON([]byte(`{ "end_points" : "test"}`), map[string]interface{}{
		"end_points": &endpoint,
	})
	g.Expect(err).To(HaveOccurred())
}
