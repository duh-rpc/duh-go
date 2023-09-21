package duh_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/harbor-pkgs/duh"
	"github.com/harbor-pkgs/duh/demo"
	"github.com/harbor-pkgs/duh/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDemoHappyPath(t *testing.T) {
	// Create a new instance of our service
	service := demo.NewService()

	// Create a new server which handles the HTTP requests for our service
	server := httptest.NewServer(&demo.Handler{Service: service})
	defer server.Close()

	// Create a new client to make RPC calls to the service via the HTTP Handler
	c := demo.NewClient(demo.ClientConfig{Endpoint: server.URL})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Test happy path JSON request and response
	{
		req := demo.SayHelloRequest{
			Name: "Admiral Thrawn",
		}
		var resp demo.SayHelloResponse
		assert.NoError(t, c.SayHello(ctx, &req, &resp))
		assert.Equal(t, "Hello, Admiral Thrawn", resp.Message)
	}

	// Test happy path Protobuf request and response
	{
		req := demo.RenderPixelRequest{
			Complexity: 1024,
			Height:     2048,
			Width:      2048,
			I:          1,
			J:          1,
		}

		var resp demo.RenderPixelResponse
		assert.NoError(t, c.RenderPixel(ctx, &req, &resp))
		assert.Equal(t, int64(72), resp.Gray)
	}
}

type badTransport struct {
}

func (t *badTransport) RoundTrip(rq *http.Request) (*http.Response, error) {
	return nil, nil
}

var badTransportClient = http.Client{Transport: &badTransport{}}

func TestClientErrors(t *testing.T) {
	service := test.NewService()
	server := httptest.NewServer(&test.Handler{Service: service})
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	for _, tt := range []struct {
		req     *test.ErrorsRequest
		details map[string]string
		conf    test.ClientConfig
		error   string
		name    string
		msg     string
		code    int
	}{
		{
			name:  "fail to marshal protobuf request",
			error: "Client Error: while marshaling request payload: string field contains invalid UTF-8",
			conf:  test.ClientConfig{Endpoint: server.URL},
			req: &test.ErrorsRequest{
				Case: string([]byte{0x80, 0x81}),
			},
			code: duh.CodeClientError,
		},
		{
			name:  "fail to create request",
			error: "Client Error: net/http: invalid method \"invalid method\"",
			conf:  test.ClientConfig{Endpoint: ""},
			req:   &test.ErrorsRequest{Case: test.CaseInvalidMethod},
			code:  duh.CodeClientError,
		},
		{
			name:    "fail to create send request",
			error:   "Client Error: Post \"/v1/test.errors\": unsupported protocol scheme \"\"",
			details: map[string]string{"http.method": "POST", "http.url": "/v1/test.errors"},
			conf:    test.ClientConfig{Endpoint: ""},
			req:     &test.ErrorsRequest{},
			code:    duh.CodeClientError,
		},
		{
			name: "fail to create request",
			error: fmt.Sprintf("Client Error: Post \"%s/v1/test.errors\": http: RoundTripper "+
				"implementation (*duh_test.badTransport) "+"returned a nil *Response with a nil error", server.URL),
			details: map[string]string{
				duh.DetailsHttpUrl:    fmt.Sprintf("%s/v1/test.errors", server.URL),
				duh.DetailsHttpMethod: "POST",
			},
			conf: test.ClientConfig{Endpoint: server.URL, Client: &badTransportClient},
			req:  &test.ErrorsRequest{},
			code: duh.CodeClientError,
		},
		{
			name:  "fail to read body of response",
			error: "Transport Error: while reading response body: unexpected EOF",
			details: map[string]string{
				duh.DetailsHttpUrl:    fmt.Sprintf("%s/v1/test.errors", server.URL),
				duh.DetailsHttpMethod: "POST",
				duh.DetailsHttpStatus: "200 OK",
			},
			req:  &test.ErrorsRequest{Case: test.CaseClientIOError},
			conf: test.ClientConfig{Endpoint: server.URL},
			code: duh.CodeTransportError,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c := test.NewClient(tt.conf)
			err := c.TestErrors(ctx, tt.req)
			var e duh.Error
			require.True(t, errors.As(err, &e))
			assert.Equal(t, tt.error, e.Error())
			assert.Equal(t, tt.msg, e.Message())
			assert.Equal(t, tt.code, e.Code())
			for k, v := range tt.details {
				require.Contains(t, e.Details(), k)
				assert.Contains(t, e.Details()[k], v)
			}
		})
	}
}

// TODO: Test no such method case

// TODO: Setup Error cases for tests
// TODO: Update the benchmark tests
//if err := c.SayHello(ctx, &req, &resp); err != nil {
//assert.NoError(t, err)
//var de duh.Error
//if errors.As(err, &de) {
//
//	//de.Details()
//	//de.Error()
//	////msg := `HTTP failed on 'GET https://example.com' (X-Account-Id: '', X-Other-Thing: '') with '404' message 'Not Found'`
//	//msg := `GET https://example.com failed with 'Not Found' message 'Fido is not in the Pet Shop'`
//	//msg = `GET https://example.com failed with 'Request Failed' message 'while reading response body: EOF'`
//	//msg = `GET https://example.com failed with 'Request Failed' message 'while parsing response body: expected {'`
//	//fmt.Printf(msg)
//
//}

// TODO: Start Testing the client!

// Is this a retryable error?
// Is this an infra error?
// Is this a failure?
// Can I tell the diff between an infra error and an error from the service?
