package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/harbor-pkgs/duh"
	"net/http"
)

//var bufferPool = sync.Pool{
//	New: func() interface{} { return new(bytes.Buffer) },
//}

type SayHelloResponse struct {
	Message string `json:"message"`
}

type SayHelloRequest struct {
	Name string `json:"name"`
}

type Service struct{}

func (i *Service) SayHello(ctx context.Context, req *SayHelloRequest, resp *SayHelloResponse) error {
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
		// TODO: Validate the payload is valid (Shouldn't the Implementation do this?)
		// TODO: Nothing stopping you from Reading the Headers to determine what todo with the payload (GroupCache?)

		// Reuse buffers for memory allocation efficiency.
		//b := bufferPool.Get().(*bytes.Buffer)
		//defer bufferPool.Put(b)
		//b.Reset()
		//
		//_, err := io.Copy(b, r.Body)
		//if err != nil {
		//	duh.ReplyMsg(w, r, duh.CodeInternalError, nil, err.Error())
		//	return
		//}
		// TODO: Support both JSON and Protobuf

		var req SayHelloRequest
		var b []byte
		err := json.Unmarshal(b, &req)
		if err != nil {
			duh.ReplyMsg(w, r, duh.CodeBadRequest, nil, err.Error())
			return
		}

		var resp SayHelloResponse
		if err := h.service.SayHello(r.Context(), &req, &resp); err != nil {
			// If SayHello ONLY returns duh.ServerError then ServerReply will do the right thing
			// If SayHello() returns a non duh.ServerError then it will return an 'Internal Error'
			// to the caller.
			duh.ReplyError(w, r, err)
			return
		}
	}
}

func main() {

}
