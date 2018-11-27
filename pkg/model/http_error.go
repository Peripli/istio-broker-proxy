package model

import "net/http"

type HttpError struct {
	Message     string `json:"error"`
	Description string `json:"description"`
	Status      int    `json:"-"`
}

func HttpErrorFromError(err error) *HttpError {
	switch t := err.(type) {
	case *HttpError:
		return t
	default:
		return &HttpError{err.Error(), "", http.StatusInternalServerError}
	}
}

func (e HttpError) Error() string {
	return e.Message
}
