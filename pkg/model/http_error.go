package model

import (
	"encoding/json"
	"fmt"
	"net/http"
)

//HTTPError represents an error in the HTTP protocol
type HTTPError struct {
	ErrorMsg    string `json:"error"`
	Description string `json:"description"`
	StatusCode  int    `json:"-"`
}

//HTTPErrorFromError converts an error to an error of type HTTPError
func HTTPErrorFromError(err error, statusCode int) *HTTPError {
	switch t := err.(type) {
	case *HTTPError:
		return t
	case HTTPError:
		return &t
	default:
		switch statusCode {
		case http.StatusBadGateway:
			return &HTTPError{"BadGateway", err.Error(), statusCode}
		default:
			return &HTTPError{"InternalServerError", err.Error(), statusCode}
		}
	}
}

//HTTPErrorFromResponse returns an HTTPError if the provided request was unsuccessful
func HTTPErrorFromResponse(statusCode int, body []byte, url string, method string, contentType string) error {
	okResponse := statusCode/100 == 2
	if !okResponse {
		var httpError HTTPError
		if contentType == "application/json" {
			err := json.Unmarshal(body, &httpError)
			if err != nil {
				return &HTTPError{StatusCode: statusCode, ErrorMsg: "InvalidJSON", Description: fmt.Sprintf("invalid JSON '%s': from call to %s %s", string(body), method, url)}
			}
		} else {
			return &HTTPError{StatusCode: statusCode, ErrorMsg: fmt.Sprintf("%s to %s failed", method, url), Description: string(body)}
		}

		httpError.StatusCode = statusCode
		httpError.Description = httpError.Description + fmt.Sprintf(": from call to %s %s", method, url)
		return &httpError
	}
	return nil
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("error: '%s', description: '%s'", e.ErrorMsg, e.Description)
}
