package duh_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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
		if err := duh.Unmarshal(r, &req); err != nil {
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

	// TODO: Implement a Client and call the server
}
