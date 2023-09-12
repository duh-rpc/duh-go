package main

import (
	"context"
	"fmt"
	"github.com/harbor-pkgs/duh"
	v1 "github.com/harbor-pkgs/duh/proto/v1"
	"net/http"
)

type Service struct{}

func (i *Service) SayHello(ctx context.Context, req *v1.SayHelloRequest, resp *v1.SayHelloResponse) error {
	// TODO: Validate the payload is valid (If the name isn't capitalized, we should reject it for giggles)
	resp.Message = fmt.Sprintf("Hello, %s", req.Name)
	return nil
}

type Handler struct {
	service Service
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {

	// TODO: Middleware
	// TODO: Authentication
	// TODO: Authorization
	// TODO: Max Read Limit Middleware
	// TODO: Rate Limit Middleware

	// No need for fancy routers, a switch case is performant and simple.
	case "/v1/say.hello":
		// TODO: Verify the authenticated user can access this endpoint
		// TODO: Nothing stopping you from Reading the Headers to determine what todo with the payload (GroupCache?)
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
		duh.Respond(w, r, duh.CodeOK, &resp)
		return
	}
}

func main() {

}
