package duh_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/harbor-pkgs/duh"
	v1 "github.com/harbor-pkgs/duh/proto/v1"
)

type HelloService struct{}

func (i *HelloService) SayHello(ctx context.Context, req *v1.SayHelloRequest, resp *v1.SayHelloResponse) error {
	resp.Message = fmt.Sprintf("Hello, %s", req.Name)
	return nil
}

type Handler struct {
	service HelloService
}

//err := json.Unmarshal(b, &req)
//var b []byte

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	// No need for fancy routers, a switch case is performant and simple.
	case "/v1/say.hello":
		var req v1.SayHelloRequest
		if err := duh.Unmarshal(r, &req); err != nil {
			duh.ReplyMsg(w, r, duh.CodeBadRequest, nil, err.Error())
			return
		}
		var resp v1.SayHelloResponse
		if err := h.service.SayHello(r.Context(), &req, &resp); err != nil {
			// If SayHello ONLY returns duh.ServerError then ServerReply will do the right thing
			// If SayHello() returns a non duh.ServerError then it will return an 'Internal Error'
			// to the caller.
			duh.ReplyError(w, r, err)
			return
		}
	}
}

func TestServer(t *testing.T) {

}
