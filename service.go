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
)

var memory = sync.Pool{
	New: func() interface{} { return bytes.NewBuffer(make([]byte, 2048)) },
}

// ReadRequest reads the given http.Request body []byte into the given proto.Message.
// The provided message must be mutable (e.g., a non-nil pointer to a message).
// It also handles content negotiation via the 'Content-Type' header provided in the http.Request headers
func ReadRequest(r *http.Request, m proto.Message) error {
	b := memory.Get().(*bytes.Buffer)
	defer memory.Put(b)
	b.Reset()

	_, err := io.Copy(b, r.Body)
	if err != nil {
		return NewErrService(CodeBadRequest, "", err, nil)
	}

	// Ignore multiple mime types separated by comma ',' or mime type parameters separated by semicolon ';'
	mimeType := TrimSuffix(r.Header.Get("Content-Type"), ";,")

	switch strings.TrimSpace(strings.ToLower(mimeType)) {
	case "", ContentTypeJSON:
		if err := json.Unmarshal(b.Bytes(), m); err != nil {
			return NewErrService(CodeBadRequest, "", err, nil)
		}
		return nil
	case ContentTypeProtoBuf:
		// TODO: Implement protobuf next
	}
	return NewErrService(CodeBadRequest,
		fmt.Sprintf("Content-Type header '%s' is invalid format or unrecognized content type",
			r.Header.Get("Content-Type")), nil, nil)
}

// ReplyWithCode replies to the request with the specified message and status code
func ReplyWithCode(w http.ResponseWriter, r *http.Request, code int, details map[string]string, msg string) {
	Reply(w, r, code, &v1.Reply{
		Code:    int32(code),
		Message: msg,
		Details: details,
	})
}

// ReplyError replies to the request with the error provided. If 'err' satisfies the Error interface,
// then it will return the code and message provided by the Error. If 'err' does not satisfy the Error
// it will then return a status of CodeInternalError with the err.Reply() as the message.
func ReplyError(w http.ResponseWriter, r *http.Request, err error) {
	var re Error
	if errors.As(err, &re) {
		Reply(w, r, re.StatusCode(), re.ProtoMessage())
		return
	}
	// If err has no Error in the error chain, then reply with CodeInternalError and the message
	// provided.
	ReplyWithCode(w, r, CodeInternalError, nil, err.Error())
}

// Reply responds to a request with the specified protobuf message and status code.
// Reply() provides content negotiation for protobuf if the request has the 'Accept' header set.
// If no 'Accept' header was provided, Reply() will marshall the proto.Message into JSON.
func Reply(w http.ResponseWriter, r *http.Request, code int, resp proto.Message) {
	// Ignore multiple mime types separated by comma ',' or mime type parameters separated by semicolon ';'
	mimeType := TrimSuffix(r.Header.Get("Accept"), ";,")

	switch strings.TrimSpace(strings.ToLower(mimeType)) {
	case "", ContentTypeJSON:
		b, err := json.Marshal(resp)
		if err != nil {
			ReplyWithCode(w, r, CodeInternalError, nil, err.Error())
			return
		}
		w.Header().Set("Content-Type", ContentTypeJSON)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(code)
		_, _ = w.Write(b)
	case ContentTypeProtoBuf:
		// TODO: Implement protobuf
		fallthrough
	default:
		r.Header.Set("Accept", ContentTypeJSON)
		ReplyWithCode(w, r, CodeBadRequest, nil, fmt.Sprintf("Accept header '%s' is invalid format "+
			"or unrecognized content type, only [%s] are supported by this method",
			mimeType, strings.Join(SupportedMimeTypes, ",")))
	}
}

// TrimSuffix trims everything after the first separator is found
func TrimSuffix(s, sep string) string {
	if i := strings.IndexAny(s, sep); i >= 0 {
		return s[:i]
	}
	return s
}

const (
	ContentTypeProtoBuf = "application/protobuf"
	ContentTypeJSON     = "application/json"
)

var SupportedMimeTypes = []string{ContentTypeJSON, ContentTypeProtoBuf}