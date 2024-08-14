package ahttp

import (
	"fmt"
	"net/http"
)

var (
	ErrBadRequest           = NewHTTPError(http.StatusBadRequest)
	ErrNotFound             = NewHTTPError(http.StatusNotFound)
	ErrUnsupportedMediaType = NewHTTPError(http.StatusUnsupportedMediaType)
)

func NewHTTPError(code int, message ...interface{}) *HTTPError {
	err := &HTTPError{
		Code:    code,
		Message: http.StatusText(code),
	}
	if len(message) > 0 {
		err.Message = message[0]
	}
	return err
}

type HTTPError struct {
	Code     int         `json:"-"`
	Message  interface{} `json:"message"`
	Internal error       `json:"-"`
}

func (e *HTTPError) Error() string {
	if e.Internal == nil {
		return fmt.Sprintf("code=%d, message=%v", e.Code, e.Message)
	}
	return fmt.Sprintf("code=%d, message=%v, internal=%v", e.Code, e.Message, e.Internal)
}

func (e *HTTPError) SetInternal(err error) *HTTPError {
	e.Internal = err
	return e
}
