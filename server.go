package duh

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	v1 "github.com/harbor-pkgs/duh/proto/v1"
	json "google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoiface"
)

var bufferPool = sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}

type ServerError interface {
	ProtoMessage() proto.Message
	StatusCode() int
}

type ErrBadRequest struct {
	Details map[string]string
	Message string
	Code    int
}

func (r ErrBadRequest) ProtoMessage() proto.Message {
	return &v1.Error{
		Code:    CodeBadRequest,
		Details: r.Details,
		Message: r.Message,
	}
}

func (r ErrBadRequest) StatusCode() int {
	return CodeBadRequest
}

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

// ReplyMsg replies to the request with the specified message and status code
func ReplyMsg(w http.ResponseWriter, r *http.Request, c int, d map[string]string, m string) {
	Reply(w, r, c, &v1.Error{
		Code:    int32(c),
		Message: m,
		Details: d,
	})
}

// ReplyMsgf is identical to ReplyMsg, but it accepts a format specifier and arguments for that satisfies the format
func ReplyMsgf(w http.ResponseWriter, r *http.Request, c int, d map[string]string, f string, a ...any) {
	ReplyMsg(w, r, c, d, fmt.Sprintf(f, a...))
}

// ReplyError replies to the request with the error provided. If 'err' satisfies the ServerError interface,
// then it will return the code and message provided by the ServerError. If 'err' does not satisfy the ServerError
// it will then return a status of CodeInternalError with the err.Error() as the message.
func ReplyError(w http.ResponseWriter, r *http.Request, err error) {
	var se ServerError
	if errors.As(err, &se) {
		Reply(w, r, se.StatusCode(), se.ProtoMessage())
		return
	}
	// If err has no ServerError in the error chain, then reply with Internal Error and the message
	// provided.
	ReplyMsg(w, r, CodeInternalError, nil, err.Error())
}

// Reply replies to the request with the specified protobuf message and status code.
// Reply() provides content negotiation for protobuf if the request has the 'Accept' header set.
// If no 'Accept' header was provided, Reply() will marshall the proto.Message into JSON.
func Reply(w http.ResponseWriter, r *http.Request, c int, p proto.Message) {
	switch strings.ToLower(r.Header.Get("Accept")) {
	case "", "application/json":
		b, err := json.Marshal(p)
		if err != nil {
			ReplyMsg(w, r, CodeInternalError, nil, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(c)
		_, _ = w.Write(b)
	case "application/protobuf":
		// TODO: Implement protobuf
		fallthrough
	default:
		r.Header.Set("Accept", "application/json")
		ReplyMsgf(w, r, CodeBadRequest, nil, "'Accept: %s' contains no acceptable mime formats. "+
			"Server supports the following (application/json)", r.Header.Get("Accept"))
	}
}

func protoMarshal(in proto.Message) ([]byte, error) {
	body := bufferPool.Get().(*bytes.Buffer)
	body.Reset()
	defer bufferPool.Put(body)

	out, err := proto.MarshalOptions{}.MarshalState(protoiface.MarshalInput{
		Message: in.ProtoReflect(),
		Buf:     body.Bytes(),
	})
	return out.Buf, err
}
