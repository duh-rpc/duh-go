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
		return NewRequestError(CodeBadRequest, "", err, nil)
	}

	// Ignore multiple mime types separated by comma ',' or mime type parameters separated by semicolon ';'
	mt := TrimSuffix(r.Header.Get("Content-Type"), ";,")

	switch strings.TrimSpace(strings.ToLower(mt)) {
	case "", "application/json":
		if err := json.Unmarshal(b.Bytes(), m); err != nil {
			return NewRequestError(CodeBadRequest, "", err, nil)
		}
	case "application/protobuf":
		// TODO:
	}
	return NewRequestError(CodeBadRequest,
		fmt.Sprintf("Content-Type header '%s' is invalid format or unrecognized content type",
			r.Header.Get("Content-Type")), nil, nil)
}

// ReplyWithCode replies to the request with the specified message and status code
func ReplyWithCode(w http.ResponseWriter, r *http.Request, code int, details map[string]string, msg string) {
	Respond(w, r, code, &v1.Reply{
		Code:    int32(code),
		Message: msg,
		Details: details,
	})
}

// ReplyError replies to the request with the error provided. If 'err' satisfies the ErrorInterface interface,
// then it will return the code and message provided by the ErrorInterface. If 'err' does not satisfy the ErrorInterface
// it will then return a status of CodeInternalError with the err.Respond() as the message.
func ReplyError(w http.ResponseWriter, r *http.Request, err error) {
	var re ErrorInterface
	if errors.As(err, &re) {
		Respond(w, r, re.StatusCode(), re.ProtoMessage())
		return
	}
	// If err has no ErrorInterface in the error chain, then reply with CodeInternalError and the message
	// provided.
	ReplyWithCode(w, r, CodeInternalError, nil, err.Error())
}

// Respond responds to a request with the specified protobuf message and status code.
// Respond() provides content negotiation for protobuf if the request has the 'Accept' header set.
// If no 'Accept' header was provided, Respond() will marshall the proto.Message into JSON.
func Respond(w http.ResponseWriter, r *http.Request, code int, resp proto.Message) {
	// Ignore multiple mime types separated by comma ',' or mime type parameters separated by semicolon ';'
	mt := TrimSuffix(r.Header.Get("Accept"), ";,")
	switch strings.TrimSpace(strings.ToLower(mt)) {
	case "", "application/json":
		b, err := json.Marshal(resp)
		if err != nil {
			ReplyWithCode(w, r, CodeInternalError, nil, err.Error())
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
		ReplyWithCode(w, r, CodeBadRequest, nil, fmt.Sprintf("Accept header '%s' is invalid format "+
			"or unrecognized content type, only [%s] are supported by this method",
			mt, strings.Join(SupportedMimeTypes, ",")))
	}
}

// TrimSuffix trims everything after the first separator is found
func TrimSuffix(s, sep string) string {
	if i := strings.IndexAny(s, sep); i >= 0 {
		return s[:i]
	}
	return s
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

var SupportedMimeTypes = []string{"application/json", "application/protobuf"}
