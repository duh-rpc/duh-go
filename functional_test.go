package duh_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/harbor-pkgs/duh"
	v1 "github.com/harbor-pkgs/duh/proto/v1"
)

// HelloClient is a simple service which calls the HelloService
type HelloClient struct {
}

func (c *HelloClient) SayHello(ctx context.Context, req *v1.SayHelloRequest, resp *v1.SayHelloResponse) error {
	return nil
}

func NewHelloClient() *HelloClient {
	return &HelloClient{}
}

// HelloService is a simple service implementation
type HelloService struct{}

func (h *HelloService) SayHello(ctx context.Context, req *v1.SayHelloRequest, resp *v1.SayHelloResponse) error {
	resp.Message = fmt.Sprintf("Hello, %s", req.Name)
	return nil
}

type Handler struct {
	service HelloService
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	// No need for fancy routers, a switch case is performant and simple.
	case "/v1/say.hello":
		var req v1.SayHelloRequest
		if err := duh.ReadRequest(r, &req); err != nil {
			duh.ReplyError(w, r, err)
			return
		}
		var resp v1.SayHelloResponse
		if err := h.service.SayHello(r.Context(), &req, &resp); err != nil {
			duh.ReplyError(w, r, err)
			return
		}
		//duh.WriteChunk(w, r, []byte)
		duh.Respond(w, r, duh.CodeOK, &resp)
		return
	}
}

func TestServer(t *testing.T) {
	server := httptest.NewServer(&Handler{})
	defer server.Close()

	c := NewHelloClient()
	var resp v1.SayHelloResponse

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := c.SayHello(ctx, &v1.SayHelloRequest{}, &resp); err != nil {
		var de duh.Error
		if errors.As(err, &de) {
			de.Details()
			de.Error()
			//msg := `HTTP failed on 'GET https://example.com' (X-Account-Id: '', X-Other-Thing: '') with '404' message 'Not Found'`
			msg := `GET https://example.com failed with 'Not Found' message 'Fido is not in the Pet Shop'`
			msg = `GET https://example.com failed with 'Request Failed' message 'while reading response body: EOF'`
			msg = `GET https://example.com failed with 'Request Failed' message 'while parsing response body: expected {'`
			fmt.Printf(msg)
		}

		// TODO: Start Testing the client!

		// Is this a retryable error?
		// Is this an infra error?
		// Is this a failure?
		// Can I tell the diff between an infra error and an error from the service?
	}
}
