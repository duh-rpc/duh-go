package duh

import (
	"errors"
	"fmt"

	v1 "github.com/harbor-pkgs/duh/proto/v1"
	"google.golang.org/protobuf/proto"
)

const (
	CodeOK               = 200
	CodeBadRequest       = 400
	CodeUnauthorized     = 401
	CodeRequestFailed    = 402
	CodeMethodNotAllowed = 403
	CodeConflict         = 409
	CodeClientError      = 428
	CodeTooManyRequests  = 429
	CodeInternalError    = 500
)

type ServerError interface {
	ProtoMessage() proto.Message
	StatusCode() int
}

type ErrBadRequest struct {
	Details map[string]string
	Wrapped error
}

func (e *ErrBadRequest) ProtoMessage() proto.Message {
	return &v1.Error{
		Message: e.Wrapped.Error(),
		Code:    CodeBadRequest,
		Details: e.Details,
	}
}

func (e *ErrBadRequest) StatusCode() int {
	return CodeBadRequest
}

func (e *ErrBadRequest) Error() string {
	return "Bad Request:" + e.Wrapped.Error()
}

type ErrUnauthorized struct {
	Details map[string]string
	Wrapped error
}

func (e *ErrUnauthorized) ProtoMessage() proto.Message {
	return &v1.Error{
		Message: e.Wrapped.Error(),
		Code:    CodeUnauthorized,
		Details: e.Details,
	}
}

func (e *ErrUnauthorized) StatusCode() int {
	return CodeUnauthorized
}

func (e *ErrUnauthorized) Error() string {
	return "Unauthorized: " + e.Wrapped.Error()
}

type ErrInternal struct {
	Details map[string]string
	Err     error
}

func (e *ErrInternal) ProtoMessage() proto.Message {
	return &v1.Error{
		Message: e.Err.Error(),
		Code:    CodeInternalError,
		Details: e.Details,
	}
}

func (e *ErrInternal) Unwrap() error {
	return e.Err
}

func (e *ErrInternal) Is(target error) bool {
	_, ok := target.(*ErrInternal)
	return ok
}

func (e *ErrInternal) StatusCode() int {
	return CodeInternalError
}

func (e *ErrInternal) Error() string {
	return "Internal Error:" + e.Err.Error()
}

// NewError returns a new proto error of type *v1.Error
func NewError(code int, err error, details map[string]string) error {
	switch code {
	case CodeBadRequest:
		return &ErrBadRequest{
			Details: details,
			Wrapped: err,
		}
		// TODO: Add all Error Codes here
	default:
		return &ErrInternal{
			Details: details,
			Err:     err,
		}
	}
}

// Error returns an error of the given code and message
func Error(code int, details map[string]string, msg string) error {
	return NewError(code, errors.New(msg), details)
}

// Errorf returns a Status with a formatted error message (Supports %w)
func Errorf(code int, details map[string]string, format string, a ...interface{}) error {
	return NewError(code, fmt.Errorf(format, a...), details)
}
