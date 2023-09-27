package test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/duh-rpc/duh-go"
	"google.golang.org/protobuf/proto"
	"net/http"
)

// Client is a simple client that calls the Service
type Client struct {
	*duh.Client
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
		Client: &duh.Client{
			Client: conf.Client,
		},
		endpoint: conf.Endpoint,
	}
}

// TestErrors is used in test suite to test error handling
func (c *Client) TestErrors(ctx context.Context, req *ErrorsRequest) error {
	payload, err := proto.Marshal(req)
	if err != nil {
		return duh.NewClientError(fmt.Errorf("while marshaling request payload: %w", err), nil)
	}

	m := http.MethodPost
	if req.Case == CaseInvalidMethod {
		m = "invalid method"
	}

	r, err := http.NewRequestWithContext(ctx, m,
		fmt.Sprintf("%s/%s", c.endpoint, "v1/test.errors"), bytes.NewReader(payload))
	if err != nil {
		return duh.NewClientError(err, nil)
	}

	r.Header.Set("Content-Type", duh.ContentTypeProtoBuf)
	var resp proto.Message
	return c.Do(r, resp)
}
