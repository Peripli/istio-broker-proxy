package model

type Service struct {
	Name                 string
	Plans                []Plan
	AdditionalProperties AdditionalProperties
}

func (s *Service) UnmarshalJSON(b []byte) error {
	return s.AdditionalProperties.UnmarshalJSON(b, map[string]interface{}{"name": &s.Name, "plans": &s.Plans})
}

func (s Service) MarshalJSON() ([]byte, error) {
	return s.AdditionalProperties.MarshalJSON(map[string]interface{}{"name": &s.Name, "plans": &s.Plans})
}
