package duh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	v1 "github.com/duh-rpc/duh-go/proto/v1"
	json "google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type HTTPClient struct {
	Client *http.Client
}

const (
	DetailsHttpCode   = "http.code"
	DetailsHttpUrl    = "http.url"
	DetailsHttpMethod = "http.method"
	DetailsHttpStatus = "http.status"
	DetailsHttpBody   = "http.body"
)

// DoWithRetry is identical to Do() except it will retry using the default retry.UntilSuccess which
// will retry all requests that return one of the following status codes.
//
//	500 - Internal Error
//	502 - Bad Gateway
//	503 - Service Unavailable
//	505 - Gateway Timeout
//	429 - Too Many Requests
//
// On 429 if the server provides a reset-time, DoWithRetry will calculate the appropriate retry time and
// sleep until that time occurs or until the context is canceled.
func (c *HTTPClient) DoWithRetry(ctx context.Context, req *http.Request, out proto.Message) error {
	return c.Do(req, out)
	// TODO: Finish retry package and enable this
	//return retry.On(ctx, retry.UntilSuccess, func(ctx context.Context, i int) error {
	//	return c.do(req, out)
	//})
}

// Do calls http.Client.Do() and un-marshals the response into the proto struct passed.
// In the case of unexpected request or response errors, Do will return *duh.ClientError
// with as much detail as possible.
func (c *HTTPClient) Do(req *http.Request, out proto.Message) error {
	// Preform the HTTP call
	resp, err := c.Client.Do(req)
	if err != nil {
		return &ClientError{
			details: map[string]string{
				DetailsHttpUrl:    req.URL.String(),
				DetailsHttpMethod: req.Method,
			},
			code: CodeClientError,
			err:  err,
		}
	}
	defer func() { _ = resp.Body.Close() }()

	body := memory.Get().(*bytes.Buffer)
	body.Reset()
	defer memory.Put(body)

	// Copy the response into a buffer
	if _, err = io.Copy(body, resp.Body); err != nil {
		return &ClientError{
			err: fmt.Errorf("while reading response body: %v", err),
			details: map[string]string{
				DetailsHttpUrl:    req.URL.String(),
				DetailsHttpMethod: req.Method,
				DetailsHttpStatus: resp.Status,
			},
			code: CodeTransportError,
		}
	}

	// If we get a code that is not a known DUH code, then don't attempt to un-marshal,
	// instead read the body and return an error
	if !IsReplyCode(resp.StatusCode) {
		return ClientErrorFromBody(req, resp, body.Bytes(), nil)
	}

	// Handle content negotiation and un-marshal the response
	mt := TrimSuffix(resp.Header.Get("Content-Type"), ";,")
	switch strings.TrimSpace(strings.ToLower(mt)) {
	case "", ContentTypeJSON:
		return c.handleJSONResponse(req, resp, body.Bytes(), out)
	case ContentTypeProtoBuf:
		return c.handleProtobufResponse(req, resp, body.Bytes(), out)
	}
	return nil
}

func (c *HTTPClient) handleJSONResponse(req *http.Request, resp *http.Response, body []byte, out proto.Message) error {
	if resp.StatusCode != CodeOK {
		var reply v1.Reply
		if err := json.Unmarshal(body, &reply); err != nil {
			// Assume the body is not a Reply structure because
			// the server is not respecting the spec.

			// TODO: Ensure this error turns into a non standard error, such that clients users can
			//  clearly identify this was a non standard 404 or whatever.
			return ClientErrorFromBody(req, resp, body, nil)
		}
		return ClientErrorFromReply(req, resp, &reply)
	}

	if err := json.Unmarshal(body, out); err != nil {
		return NewErrService(CodeClientError, "",
			fmt.Errorf("while parsing response body '%s': %w", body, err), nil)
	}
	return nil
}

func (c *HTTPClient) handleProtobufResponse(req *http.Request, resp *http.Response, body []byte, out proto.Message) error {
	if resp.StatusCode != CodeOK {
		var reply v1.Reply
		if err := proto.Unmarshal(body, &reply); err != nil {
			return ClientErrorFromBody(req, resp, body, nil)
		}
		return ClientErrorFromReply(req, resp, &reply)
	}

	if err := proto.Unmarshal(body, out); err != nil {
		return NewErrService(CodeClientError, "",
			fmt.Errorf("while parsing response body '%s': %w", body, err), nil)
	}
	return nil
}

func ClientErrorFromReply(req *http.Request, resp *http.Response, reply *v1.Reply) error {
	details := map[string]string{
		DetailsHttpCode:   fmt.Sprintf("%d", resp.StatusCode),
		DetailsHttpUrl:    req.URL.String(),
		DetailsHttpStatus: resp.Status,
		DetailsHttpMethod: req.Method,
	}

	for k, v := range reply.Details {
		details[k] = v
	}

	// TODO: Handle CodeTooManyRequests, include a way to easily get those retry values
	//  so retry.On() can get them.
	return &ClientError{
		code:    int(reply.Code),
		msg:     reply.Message,
		details: details,
	}
}

func ClientErrorFromBody(req *http.Request, resp *http.Response, body []byte, err error) error {
	return &ClientError{
		details: map[string]string{
			DetailsHttpCode:   fmt.Sprintf("%d", resp.StatusCode),
			DetailsHttpBody:   fmt.Sprintf("%s", body),
			DetailsHttpUrl:    req.URL.String(),
			DetailsHttpStatus: resp.Status,
			DetailsHttpMethod: req.Method,
		},
		code: resp.StatusCode,
		// TODO: If err is nil and msg is empty, then report it as 'Unrecognized reply'
		err: err,
	}
}

func NewClientError(code int, err error, details map[string]string) error {
	return &ClientError{
		details: details,
		code:    code,
		err:     err,
	}
}
