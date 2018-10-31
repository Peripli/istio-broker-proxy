package model

type Service struct {
	Id                   string `json:"id"`
	AdditionalProperties AdditionalProperties
}

func (s *Service) UnmarshalJSON(b []byte) error {
	return s.AdditionalProperties.UnmarshalJSON(b, map[string]interface{}{"id": &s.Id})
}

func (s Service) MarshalJSON() ([]byte, error) {
	return s.AdditionalProperties.MarshalJSON(map[string]interface{}{"id": &s.Id})
}
