package model

import "encoding/json"

type Plan struct {
	MetaData             map[string]json.RawMessage `json:"metadata"`
	AdditionalProperties AdditionalProperties
}

func (p *Plan) UnmarshalJSON(b []byte) error {
	return p.AdditionalProperties.UnmarshalJSON(b, map[string]interface{}{"metadata": &p.MetaData})
}

func (p *Plan) MarshalJSON() ([]byte, error) {
	return p.AdditionalProperties.MarshalJSON(map[string]interface{}{"metadata": &p.MetaData})
}
