package model

type Service struct {
	Name                 string `json:"name"`
	AdditionalProperties AdditionalProperties
}

func (s *Service) UnmarshalJSON(b []byte) error {
	return s.AdditionalProperties.UnmarshalJSON(b, map[string]interface{}{"name": &s.Name})
}

func (s Service) MarshalJSON() ([]byte, error) {
	return s.AdditionalProperties.MarshalJSON(map[string]interface{}{"name": &s.Name})
}
