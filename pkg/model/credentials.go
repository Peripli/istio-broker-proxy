package model

import "encoding/json"

type Credentials struct {
	AdditionalProperties map[string]json.RawMessage
	Endpoints            []Endpoint
}

func (credentials *Credentials) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, &credentials.AdditionalProperties); err != nil {
		return err
	}
	err := removeProperty(credentials.AdditionalProperties, "end_points", &credentials.Endpoints)
	if err != nil {
		return err
	}
	return nil
}

func (credentials Credentials) MarshalJSON() ([]byte, error) {
	properties := clone(credentials.AdditionalProperties)
	if len(credentials.Endpoints) != 0 {
		addProperty(properties, "end_points", credentials.Endpoints)
	}

	return json.Marshal(properties)
}
