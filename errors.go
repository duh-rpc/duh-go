package duh

import (
	"net/http"

	v1 "github.com/harbor-pkgs/duh/proto/v1"
	"google.golang.org/protobuf/proto"
)

const (
	CodeOK               = 200
	CodeBadRequest       = 400
	CodeUnauthorized     = 401
	CodeMethodNotAllowed = 403
	CodeNotFound         = 404
	CodeConflict         = 409
	CodeTooManyRequests  = 429
	CodeClientError      = 452
	CodeRequestFailed    = 453
	CodeInternalError    = 500
	CodeNotImplemented   = 501
	CodeTransportError   = 512
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
	case CodeNotFound:
		return "Not Found"
	case CodeConflict:
		return "Conflict"
	case CodeClientError:
		return "Client Respond"
	case CodeTooManyRequests:
		return "Too Many Requests"
	case CodeInternalError:
		return "Internal Server Error"
	case CodeNotImplemented:
		return "Not Implemented"
	case CodeTransportError:
		return "Transport Error"
	default:
		return http.StatusText(code)
	}
}

func IsReplyCode(code int) bool {
	switch code {
	case CodeOK, CodeBadRequest, CodeUnauthorized, CodeRequestFailed, CodeMethodNotAllowed,
		CodeNotFound, CodeConflict, CodeClientError, CodeTooManyRequests, CodeInternalError,
		CodeNotImplemented, CodeTransportError:
		return true
	}
	return false
}

type Error interface {
	// ProtoMessage Creates v1.Reply protobuf from this Error
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

var _ Error = (*ErrService)(nil)
var _ Error = (*ErrClient)(nil)

type ErrService struct {
	details map[string]string
	msg     string
	err     error
	code    int
}

// NewErrService returns a new ErrService.
// Server Implementations should use this to respond to requests with an error.
func NewErrService(code int, msg string, err error, details map[string]string) error {
	return &ErrService{
		details: details,
		code:    code,
		msg:     msg,
		err:     err,
	}
}

func (e *ErrService) ProtoMessage() proto.Message {
	if e.err != nil && e.msg == "" {
		e.msg = e.err.Error()
	}
	return &v1.Reply{
		Code:    int32(e.code),
		Details: e.details,
		Message: e.msg,
	}
}

func (e *ErrService) StatusCode() int {
	return e.code
}

func (e *ErrService) Message() string {
	return e.msg
}

func (e *ErrService) Error() string {
	return CodeText(e.code) + ":" + e.err.Error()
}

func (e *ErrService) Details() map[string]string {
	return e.details
}

type ErrClient struct {
	details map[string]string
	msg     string
	err     error
	code    int
}

func (e *ErrClient) ProtoMessage() proto.Message {
	if e.err != nil && e.msg == "" {
		e.msg = e.err.Error()
	}
	return &v1.Reply{
		Code:    int32(e.code),
		Details: e.details,
		Message: e.msg,
	}
}

func (e *ErrClient) StatusCode() int {
	return e.code
}

func (e *ErrClient) Message() string {
	return e.msg
}

func (e *ErrClient) Error() string {
	return CodeText(e.code) + ":" + e.err.Error()
}

func (e *ErrClient) Details() map[string]string {
	return e.details
}

//// Error returns an error of the given code and message.
//// You should use this when reporting an error that did not originate from an error of type `error`
//func Error(code int, details map[string]string, msg string) error {
//	return NewErrService(code, "", errors.New(msg), details)
//}
//
//// Errorf returns a Status with a formatted error message (Supports %w)
//func Errorf(code int, details map[string]string, format string, a ...interface{}) error {
//	return NewErrService(code, "", fmt.Errorf(format, a...), details)
//}

// IsRetryable returns true if any error in the chain is of type Error, and is one
// of the following codes [429,500,502,503,504]
func IsRetryable(err error) bool {
	// TODO
	return false
}
