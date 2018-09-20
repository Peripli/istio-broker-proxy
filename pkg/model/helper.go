package model

import (
	"encoding/json"
	"fmt"
)

func addProperty(additionalProperties map[string]json.RawMessage, key string, data interface{}) {
	rawData, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Sprintf("Error in marshal %v", err))
	}
	additionalProperties[key] = json.RawMessage(rawData)

}

func removeProperty(additionalProperties map[string]json.RawMessage, key string, data interface{}) error {
	rawData := additionalProperties[key]
	if rawData != nil {
		if err := json.Unmarshal(rawData, &data); err != nil {
			return err
		}
		delete(additionalProperties, key)
	}
	return nil
}

func clone(original map[string]json.RawMessage) map[string]json.RawMessage {
	copied := make(map[string]json.RawMessage)

	for key, value := range original {
		copied[key] = value
	}
	return copied
}
