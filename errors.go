package duh

import (
	"fmt"
	"github.com/harbor-pkgs/duh/proto/v1"
)

// NewError returns a new proto error of type *v1.Error
func NewError(c int64, msg string, details map[string]string) *v1.Error {
	return &v1.Error{
		Code:    c,
		Message: msg,
		Details: details,
	}
}

// Error returns an error of the given code and message
func Error(c int64, msg string) error {
	return NewError(c, msg, nil)
}

// Errorf returns a Status with a formatted error message
func Errorf(c int64, format string, a ...interface{}) error {
	return Error(c, fmt.Sprintf(format, a...))
}

// TODO: Might end up using error.IsError() and only care about the actual type in the handler with
//  Convert to response type method of some sort.
