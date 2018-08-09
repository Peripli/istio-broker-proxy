package credentials

import (
. "github.com/onsi/gomega"
"testing"
)

func TestUpdateEmptyContent(t *testing.T) {
	g := NewGomegaWithT(t)

	_, err := Update([]byte("{}"))

	g.Expect(err).Should(HaveOccurred())
}

func TestUpdateSimpleExample(t *testing.T) {
	g := NewGomegaWithT(t)

	response, err := Update([]byte(exampleRequest))

	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(string(response)).To(haveTheEndpointMapping("appnethost", "9876"))
}
