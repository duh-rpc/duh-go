package demo

import (
	"github.com/harbor-pkgs/duh"
	"net/http"
)

type Handler struct {
	Service *Service
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// TODO: Middleware
	// TODO: Authentication
	// TODO: Authorization
	// TODO: Max Read Limit Middleware
	// TODO: Rate Limit Middleware

	// No need for fancy routers, a switch case is performant and simple.
	switch r.URL.Path {
	case "/v1/say.hello":
		h.handleSayHello(w, r)
		return
	case "/v1/render.pixel":
		h.handleRenderPixel(w, r)
		return
	}
}

func (h *Handler) handleSayHello(w http.ResponseWriter, r *http.Request) {
	// TODO: Verify the authenticated user can access this endpoint
	var req SayHelloRequest
	if err := duh.ReadRequest(r, &req); err != nil {
		duh.ReplyError(w, r, err)
		return
	}
	var resp SayHelloResponse
	if err := h.Service.SayHello(r.Context(), &req, &resp); err != nil {
		duh.ReplyError(w, r, err)
		return
	}
	duh.Reply(w, r, duh.CodeOK, &resp)
}

func (h *Handler) handleRenderPixel(w http.ResponseWriter, r *http.Request) {
	var req RenderPixelRequest
	if err := duh.ReadRequest(r, &req); err != nil {
		duh.ReplyError(w, r, err)
		return
	}
	var resp RenderPixelResponse
	if err := h.Service.RenderPixel(r.Context(), &req, &resp); err != nil {
		duh.ReplyError(w, r, err)
		return
	}
	duh.Reply(w, r, duh.CodeOK, &resp)
}
