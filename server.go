package duh

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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

// Unmarshal reads the given http.Request body []byte into the given proto.Message.
// The provided message must be mutable (e.g., a non-nil pointer to a message).
// It also handles content negotiation via the 'Content-Type' header provided in the http.Request headers
func Unmarshal(r *http.Request, m proto.Message) error {
	b := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(b)
	b.Reset()

	_, err := io.Copy(b, r.Body)
	if err != nil {
		// TODO: This is probably not an internal error, but a read limit size error, so we should return an error
		//  appropriate for that.
		// Should we use NewError() or ErrBadRequest{}. I think we should NewError() and have it return a ServerError which is easy to understand?
		return NewError(CodeBadRequest, err),
		}
	}

	switch r.Header.Get("Content-Type") {
	case "", "application/json":

		json.Unmarshal()
	case "application/protobuf":
		fallthrough
	default:

	}
}

// ReplyMsg replies to the request with the specified message and status code
func ReplyMsg(w http.ResponseWriter, r *http.Request, code int, details map[string]string, msg string) {
	Reply(w, r, code, &v1.Error{
		Code:    int32(code),
		Message: msg,
		Details: details,
	})
}

// ReplyMsgf is identical to ReplyMsg, but it accepts a format specifier and arguments for that satisfies the format
func ReplyMsgf(w http.ResponseWriter, r *http.Request, code int, details map[string]string, format string, args ...any) {
	ReplyMsg(w, r, code, details, fmt.Sprintf(format, args...))
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
func Reply(w http.ResponseWriter, r *http.Request, code int, resp proto.Message) {
	switch strings.ToLower(r.Header.Get("Accept")) {
	case "", "application/json":
		b, err := json.Marshal(resp)
		if err != nil {
			ReplyMsg(w, r, CodeInternalError, nil, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(code)
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
