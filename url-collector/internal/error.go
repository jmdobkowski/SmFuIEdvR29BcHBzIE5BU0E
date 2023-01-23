package internal

import (
	"fmt"
	"net/http"
)

type RequestError struct {
	Message string
	Status  int
}

func (err RequestError) Error() string {
	return err.Message
}

func BadRequestErrorf(format string, a ...any) error {
	return RequestError{Message: fmt.Sprintf(format, a...), Status: http.StatusBadRequest}
}

var ErrNotFound = RequestError{Message: "not found", Status: http.StatusNotFound}
