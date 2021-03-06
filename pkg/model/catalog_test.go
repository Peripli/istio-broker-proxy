package model

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
)

var exampleCatalog = `{
  "services": [{
    "name": "fake-Service",
    "id": "acb56d7c-XXXX-XXXX-XXXX-feb140a59a66",
    "description": "A fake Service.",
    "tags": ["no-sql", "relational"],
    "requires": ["route_forwarding"],
    "bindable": true,
    "instances_retrievable": true,
    "bindings_retrievable": true,
    "metadata": {
      "provider": {
        "name": "The name"
      },
      "listing": {
        "imageUrl": "http://example.com/cat.gif",
        "blurb": "Add a blurb here",
        "longDescription": "A long time ago, in a galaxy far far away..."
      },
      "displayName": "The Fake Service Broker"
    },
    "plan_updateable": true,
    "plans": [{
      "name": "fake-plan-1",
      "id": "d3031751-XXXX-XXXX-XXXX-a42377d3320e",
      "description": "Shared fake Server, 5tb persistent disk, 40 max concurrent connections.",
      "free": false,
      "metadata": {
        "max_storage_tb": 5,
        "costs":[
            {
               "amount":{
                  "usd":99.0
               },
               "unit":"MONTHLY"
            },
            {
               "amount":{
                  "usd":0.99
               },
               "unit":"1GB of messages over 20GB"
            }
         ],
        "bullets": [
          "Shared fake server",
          "5 TB storage",
          "40 concurrent connections"
        ]
      },
      "schemas": {
        "service_instance": {
          "create": {
            "parameters": {
              "$schema": "http://json-schema.org/draft-04/schema#",
              "type": "object",
              "properties": {
                "billing-account": {
                  "description": "Billing account number used to charge use of shared fake server.",
                  "type": "string"
                }
              }
            }
          },
          "update": {
            "parameters": {
              "$schema": "http://json-schema.org/draft-04/schema#",
              "type": "object",
              "properties": {
                "billing-account": {
                  "description": "Billing account number used to charge use of shared fake server.",
                  "type": "string"
                }
              }
            }
          }
        },
        "service_binding": {
          "create": {
            "parameters": {
              "$schema": "http://json-schema.org/draft-04/schema#",
              "type": "object",
              "properties": {
                "billing-account": {
                  "description": "Billing account number used to charge use of shared fake server.",
                  "type": "string"
                }
              }
            }
          }
        }
      }
    }, {
      "name": "fake-plan-2",
      "id": "0f4008b5-XXXX-XXXX-XXXX-dace631cd648",
      "description": "Shared fake Server, 5tb persistent disk, 40 max concurrent connections. 100 async.",
      "free": false,
      "metadata": {
        "max_storage_tb": 5,
        "costs":[
            {
               "amount":{
                  "usd":199.0
               },
               "unit":"MONTHLY"
            },
            {
               "amount":{
                  "usd":0.99
               },
               "unit":"1GB of messages over 20GB"
            }
         ],
        "bullets": [
          "40 concurrent connections"
        ]
      }
    }]
  }]
}`

func TestCatalogUnmarshal(t *testing.T) {
	g := NewGomegaWithT(t)
	var c Catalog
	err := json.Unmarshal([]byte(exampleCatalog), &c)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(len(c.Services)).To(Equal(1))
	g.Expect(c.Services[0].Name).To(Equal("fake-Service"))
	data, err := json.Marshal(&c)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(data)).To(MatchJSON(exampleCatalog))
}

func TestCatalogUnmarshalPlansFromCatalog(t *testing.T) {
	g := NewGomegaWithT(t)
	var catalog Catalog
	err := json.Unmarshal([]byte(exampleCatalog), &catalog)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(len(catalog.Services[0].Plans)).To(Equal(2))
}
