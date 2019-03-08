package model

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HttpError struct {
	ErrorMsg    string `json:"error"`
	Description string `json:"description"`
	StatusCode  int    `json:"-"`
}

func HttpErrorFromError(err error, statusCode int) *HttpError {
	switch t := err.(type) {
	case *HttpError:
		return t
	case HttpError:
		return &t
	default:
		switch statusCode {
		case http.StatusBadGateway:
			return &HttpError{"BadGateway", err.Error(), statusCode}
		default:
			return &HttpError{"InternalServerError", err.Error(), statusCode}
		}
	}
}

func HttpErrorFromResponse(statusCode int, body []byte, url string, method string) error {
	okResponse := statusCode/100 == 2
	if !okResponse {
		var httpError HttpError
		err := json.Unmarshal(body, &httpError)
		if err != nil {
			return &HttpError{StatusCode: statusCode, ErrorMsg: "InvalidJSON", Description: fmt.Sprintf("invalid JSON '%s': from call to %s %s", string(body), method, url)}
		}
		httpError.StatusCode = statusCode
		httpError.Description = httpError.Description + fmt.Sprintf(": from call to %s %s", method, url)
		return &httpError
	}
	return nil
}

func (e HttpError) Error() string {
	return fmt.Sprintf("error: '%s', description: '%s'", e.ErrorMsg, e.Description)
}
