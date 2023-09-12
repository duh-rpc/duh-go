package duh

import (
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

func CodeText(code int) string {
	switch code {
	case CodeOK:
		return "OK"
	case CodeBadRequest:
		return "Bad Request"
	case CodeUnauthorized:
		return "Unauthorized"
	case CodeRequestFailed:
		return "Request Failed"
	case CodeMethodNotAllowed:
		return "Method Not Allowed"
	case CodeConflict:
		return "Conflict"
	case CodeClientError:
		return "Client Respond"
	case CodeTooManyRequests:
		return "Too Many Requests"
	case CodeInternalError:
		return "Internal Respond"
	default:
		return "Unknown Code"
	}
}

type ErrorInterface interface {
	// ProtoMessage Creates v1.Reply protobuf from this ErrorInterface
	ProtoMessage() proto.Message
	// StatusCode is the HTTP status retrieved from v1.Respond.Details
	StatusCode() int
	// Error is the error message this error wrapped (Used on the server side)
	Error() string
	// Details is the details of the error retrieved from v1.Respond.Details
	Details() map[string]string
	// Message is the message retrieved from v1.Respond.Respond
	Message() string
}

var _ ErrorInterface = (*Error)(nil)

type Error struct {
	details map[string]string
	msg     string
	err     error
	code    int
}

func (e *Error) ProtoMessage() proto.Message {
	if e.err != nil && e.msg == "" {
		e.msg = e.err.Error()
	}
	return &v1.Reply{
		Code:    int32(e.code),
		Details: e.details,
		Message: e.msg,
	}
}

func (e *Error) StatusCode() int {
	return e.code
}

func (e *Error) Message() string {
	return e.msg
}

func (e *Error) Error() string {
	return CodeText(e.code) + ":" + e.err.Error()
}

func (e *Error) Details() map[string]string {
	return e.details
}

// TODO: Maybe rename this to `WrapError()` and drop the `msg` ?

// NewRequestError returns a new Error. You should use this when wrapping an error of type `error`.
func NewRequestError(code int, msg string, err error, details map[string]string) error {
	return &Error{
		details: details,
		code:    code,
		msg:     msg,
		err:     err,
	}
}

//// ErrorInterface returns an error of the given code and message.
//// You should use this when reporting an error that did not originate from an error of type `error`
//func ErrorInterface(code int, details map[string]string, msg string) error {
//	return NewRequestError(code, "", errors.New(msg), details)
//}
//
//// Errorf returns a Status with a formatted error message (Supports %w)
//func Errorf(code int, details map[string]string, format string, a ...interface{}) error {
//	return NewRequestError(code, "", fmt.Errorf(format, a...), details)
//}

// IsRetryable returns true if any error in the chain is of type ErrorInterface, and is one
// of the following codes [429,500,502,503,504]
func IsRetryable(err error) bool {
	// TODO
	return false
}
