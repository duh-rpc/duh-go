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

// SayHello sends a name to the service using JSON and the service says hello.
func (c *Client) SayHello(ctx context.Context, req *SayHelloRequest, resp *SayHelloResponse) error {
	payload, err := json.Marshal(req)
	if err != nil {
		return duh.NewClientError(duh.CodeClientError,
			fmt.Errorf("while marshaling request payload: %w", err), nil)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/%s", c.endpoint, "v1/say.hello"), bytes.NewReader(payload))
	if err != nil {
		return duh.NewClientError(duh.CodeClientError, err, nil)
	}

	// Tell the server what kind of serialization we are sending it.
	r.Header.Set("Content-Type", duh.ContentTypeJSON)

	// Do() will handle content negotiation, error handling, and un-marshal the response
	return c.Do(r, resp)
}

// RenderPixel sends a request to the service which calculates the pixel color of a Mandelbrot
// fractal at the given point in the image.
func (c *Client) RenderPixel(ctx context.Context, req *RenderPixelRequest, resp *RenderPixelResponse) error {
	payload, err := proto.Marshal(req)
	if err != nil {
		return duh.NewClientError(duh.CodeClientError,
			fmt.Errorf("while marshaling request payload: %w", err), nil)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/%s", c.endpoint, "v1/render.pixel"), bytes.NewReader(payload))
	if err != nil {
		return duh.NewClientError(duh.CodeClientError, err, nil)
	}

	r.Header.Set("Content-Type", duh.ContentTypeProtoBuf)
	return c.Do(r, resp)
}

// TestErrors is used in test suites to test error handling
func (c *Client) TestErrors(ctx context.Context, req proto.Message, resp proto.Message) error {
	payload, err := proto.Marshal(req)
	if err != nil {
		return duh.NewClientError(duh.CodeClientError,
			fmt.Errorf("while marshaling request payload: %w", err), nil)
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/%s", c.endpoint, "v1/test.errors"), bytes.NewReader(payload))
	if err != nil {
		return duh.NewClientError(duh.CodeClientError, err, nil)
	}

	r.Header.Set("Content-Type", duh.ContentTypeProtoBuf)
	return c.Do(r, resp)
}
