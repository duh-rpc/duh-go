package demo

import (
	"bytes"
	"context"
	"fmt"
	"github.com/harbor-pkgs/duh"
	json "google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"net/http"
)

// Client is a simple client that calls the Service
type Client struct {
	*duh.HTTPClient
	endpoint string
}

type ClientConfig struct {
	Endpoint string
	Client   *http.Client
}

func NewClient(conf ClientConfig) *Client {
	if conf.Client == nil {
		conf.Client = &http.Client{Transport: http.DefaultTransport}
	}
	return &Client{
		HTTPClient: &duh.HTTPClient{
			Client: conf.Client,
		},
		endpoint: conf.Endpoint,
	}
}

// SayHello sends a name to the server and expects a reply
func (c *Client) SayHello(ctx context.Context, req *SayHelloRequest, resp *SayHelloResponse) error {
	// TODO: Remove the need to return from here if not needed
	payload, err := proto.Marshal(req)
	if err != nil {
		return duh.NewClientError(duh.CodeClientError,
			fmt.Errorf("while marshaling request payload: %w", err), nil)
	}

	// TODO: Support Protobuf
	payload, err = json.Marshal(req)
	if err != nil {
		return duh.NewClientError(duh.CodeClientError,
			fmt.Errorf("while marshaling request payload: %w", err), nil)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/%s", c.endpoint, "v1/say.hello"), bytes.NewReader(payload))
	if err != nil {
		return duh.NewClientError(duh.CodeClientError, err, nil)
	}
	r.Header.Set("Content-Type", duh.ContentTypeJSON)

	// Do() will handle content negotiation, error handling, and un-marshal the response
	return c.Do(r, resp)
}

// NewService creates a new service instance
func NewService() *Service {
	return &Service{}
}

// Service is an example of a production ready service implementation
type Service struct{}

func (h *Service) SayHello(ctx context.Context, req *SayHelloRequest, resp *SayHelloResponse) error {
	// TODO: Validate the payload is valid (If the name isn't capitalized, we should reject it for giggles)
	resp.Message = fmt.Sprintf("Hello, %s", req.Name)
	return nil
}

type Handler struct {
	Service *Service
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// TODO: Middleware
	// TODO: Authentication
	// TODO: Authorization
	// TODO: Max Read Limit Middleware
	// TODO: Rate Limit Middleware

	switch r.URL.Path {
	// No need for fancy routers, a switch case is performant and simple.
	case "/v1/say.hello":
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
		return
	}
}
