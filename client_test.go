package duh_test

import (
	"context"
	"testing"
)

type DoThingRequest struct{}
type DoThingResponse struct{}

type Client struct{}

func (c *Client) DoThing(ctx context.Context, r *DoThingRequest, resp *DoThingResponse) error {

	return nil
}

func NewClient() *Client {
	return &Client{}
}

func TestClient(t *testing.T) {
	c := NewClient()
	var resp DoThingResponse

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := c.DoThing(ctx, &DoThingRequest{}, &resp); err != nil {

	}
}
