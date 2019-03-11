package model

import "encoding/json"

//Plan represents a service plan
type Plan struct {
	MetaData             map[string]json.RawMessage `json:"metadata"`
	AdditionalProperties AdditionalProperties
}

//UnmarshalJSON unmarshals to a service plan
func (p *Plan) UnmarshalJSON(b []byte) error {
	return p.AdditionalProperties.UnmarshalJSON(b, map[string]interface{}{"metadata": &p.MetaData})
}

//MarshalJSON from a service plan
func (p *Plan) MarshalJSON() ([]byte, error) {
	return p.AdditionalProperties.MarshalJSON(map[string]interface{}{"metadata": &p.MetaData})
}
