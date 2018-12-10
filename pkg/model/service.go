package model

import "encoding/json"

type Service struct {
	Name                 string
	MetaData             map[string]json.RawMessage
	AdditionalProperties AdditionalProperties
}

func (s *Service) UnmarshalJSON(b []byte) error {
	return s.AdditionalProperties.UnmarshalJSON(b, map[string]interface{}{"name": &s.Name, "metadata": &s.MetaData})
}

func (s Service) MarshalJSON() ([]byte, error) {
	return s.AdditionalProperties.MarshalJSON(map[string]interface{}{"name": &s.Name, "metadata": &s.MetaData})
}
