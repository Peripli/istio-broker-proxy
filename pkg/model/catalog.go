package model

//Catalog represents an osb-catalog
type Catalog struct {
	Services []Service `json:"services"`
}
