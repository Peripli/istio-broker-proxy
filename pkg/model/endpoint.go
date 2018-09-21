package model

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Endpoint struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func (ep *Endpoint) UnmarshalJSON(b []byte) error {
	var untyped struct {
		Host string      `json:"host"`
		Port interface{} `json:"port"`
	}
	err := json.Unmarshal(b, &untyped)
	if err != nil {
		return err
	}

	ep.Host = untyped.Host
	if untyped.Port != nil {
		portAsString := fmt.Sprintf("%v", untyped.Port)
		ep.Port, err = strconv.Atoi(portAsString)
	}
	return err
}
