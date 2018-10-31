package model

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

type AdditionalProperties map[string]json.RawMessage

func (ap *AdditionalProperties) UnmarshalJSON(b []byte, values map[string]interface{}) error {
	if err := json.Unmarshal(b, ap); err != nil {
		return err
	}
	err := removeProperties(*ap, values)
	if err != nil {
		return err
	}
	return nil
}

func (ap *AdditionalProperties) MarshalJSON(values map[string]interface{}) ([]byte, error) {
	properties := clone(*ap)
	for key, value := range values {
		if value != nil {
			val := reflect.ValueOf(value)
			if val.Kind() == reflect.Array || val.Kind() == reflect.Slice {
				if val.Len() != 0 {
					addProperty(properties, key, value)
				}
			} else {
				addProperty(properties, key, value)
			}
		}
	}
	return json.Marshal(properties)
}

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

func removeProperties(additionalProperties map[string]json.RawMessage, values map[string]interface{}) error {
	for key, value := range values {
		err := removeProperty(additionalProperties, key, value)
		if err != nil {
			return err
		}
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

func removeIntOrStringProperty(additionalProperties map[string]json.RawMessage, key string, data *int) error {
	var untyped interface{}
	err := removeProperty(additionalProperties, key, &untyped)
	if err != nil {
		return err
	}
	if untyped != nil {
		portAsString := fmt.Sprintf("%v", untyped)
		*data, err = strconv.Atoi(portAsString)
		if err != nil {
			return err
		}
	}
	return nil
}
