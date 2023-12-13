/*
Copyright 2023 Derrick J Wippler

Licensed under the MIT License, you may obtain a copy of the License at

https://opensource.org/license/mit/ or in the root of this code repo

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package duh

import (
	"fmt"
	"net/http"

	v1 "github.com/duh-rpc/duh-go/proto/v1"
	"google.golang.org/protobuf/proto"
)

const (
	CodeOK               = 200
	CodeBadRequest       = 400
	CodeUnauthorized     = 401
	CodeForbidden        = 403
	CodeNotFound         = 404
	CodeConflict         = 409
	CodeTooManyRequests  = 429
	CodeClientError      = 452
	CodeRequestFailed    = 453
	CodeInternalError    = 500
	CodeNotImplemented   = 501
	CodeTransportError   = 512
	CodeContentTypeError = 513
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
	case CodeForbidden:
		return "Forbidden"
	case CodeNotFound:
		return "Not Found"
	case CodeConflict:
		return "Conflict"
	case CodeClientError:
		return "Client Error"
	case CodeTooManyRequests:
		return "Too Many Requests"
	case CodeInternalError:
		return "Internal Service Error"
	case CodeNotImplemented:
		return "Not Implemented"
	case CodeTransportError:
		return "Transport Error"
	case CodeContentTypeError:
		return "Content Type Error"
	default:
		return http.StatusText(code)
	}
}

func IsReplyCode(code int) bool {
	switch code {
	case CodeOK, CodeBadRequest, CodeUnauthorized, CodeRequestFailed, CodeForbidden,
		CodeNotFound, CodeConflict, CodeClientError, CodeTooManyRequests, CodeInternalError,
		CodeNotImplemented, CodeTransportError, CodeContentTypeError:
		return true
	}
	return false
}

type Error interface {
	// ProtoMessage Creates v1.Reply protobuf from this Error
	ProtoMessage() proto.Message
	// Code is the code retrieved from v1.Reply.Code or the HTTP Status Code
	Code() int
	// Error is the error message this error wrapped (Used on the server side)
	Error() string
	// Details is the Details of the error retrieved from v1.Reply.details
	Details() map[string]string
	// Message is the message retrieved from v1.Reply.Reply
	Message() string
}

var _ Error = (*ServiceError)(nil)
var _ Error = (*ClientError)(nil)

// TODO: Decide if this should be public or not, I'm leaning toward not being public
type ServiceError struct {
	details map[string]string
	err     error
	code    int
}

// NewServiceError returns a new ServiceError.
// Server Implementations should use this to respond to requests with an error.
// TODO: Ensure you can get the `cause` of the error from ServiceError struct
func NewServiceError(code int, err error, details map[string]string) error {
	return &ServiceError{
		details: details,
		code:    code,
		err:     err,
	}
}

func (e *ServiceError) ProtoMessage() proto.Message {
	return &v1.Reply{
		Message: func() string {
			if e.err != nil {
				return e.err.Error()
			}
			return ""
		}(),
		Code:    int32(e.code),
		Details: e.details,
	}
}

func (e *ServiceError) Code() int {
	return e.code
}

func (e *ServiceError) Message() string {
	return e.err.Error()
}

func (e *ServiceError) Error() string {
	return CodeText(e.code) + ":" + e.err.Error()
}

func (e *ServiceError) Details() map[string]string {
	return e.details
}

// TODO: Decide if this should be public or not, I'm leaning toward not being public
type ClientError struct {
	details      map[string]string
	msg          string
	err          error
	isInfraError bool
	code         int
}

func (e *ClientError) ProtoMessage() proto.Message {
	if e.err != nil && e.msg == "" {
		e.msg = e.err.Error()
	}
	return &v1.Reply{
		Code:    int32(e.code),
		Details: e.details,
		Message: e.msg,
	}
}

func (e *ClientError) Code() int {
	return e.code
}

func (e *ClientError) Message() string {
	return e.msg
}

func (e *ClientError) Error() string {
	// If e.err is set, it means this error is from the client
	if e.err != nil {
		return CodeText(e.code) + ": " + e.err.Error()
	}

	// This means the reply is not from the service, but from the infrastructure.
	if e.isInfraError {
		return fmt.Sprintf("%s %s returned infrastructure error '%d' with body '%s'",
			e.details[DetailsHttpMethod],
			e.details[DetailsHttpUrl],
			e.code,
			e.msg,
		)
	}
	// Error is from the service
	return fmt.Sprintf("%s %s returned code '%s' and message '%s'",
		e.details[DetailsHttpMethod],
		e.details[DetailsHttpUrl],
		CodeText(e.code),
		e.msg,
	)
}

func (e *ClientError) Details() map[string]string {
	return e.details
}
