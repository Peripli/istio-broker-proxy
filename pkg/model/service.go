package model

//Service represents a named service with plans
type Service struct {
	Name                 string
	Plans                []Plan
	AdditionalProperties additionalProperties
}

//UnmarshalJSON unmarshals
func (s *Service) UnmarshalJSON(b []byte) error {
	return s.AdditionalProperties.UnmarshalJSON(b, map[string]interface{}{"name": &s.Name, "plans": &s.Plans})
}

//MarshalJSON marshals
func (s Service) MarshalJSON() ([]byte, error) {
	return s.AdditionalProperties.MarshalJSON(map[string]interface{}{"name": &s.Name, "plans": &s.Plans})
}
