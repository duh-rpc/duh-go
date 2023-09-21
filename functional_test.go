package duh_test

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/harbor-pkgs/duh/proto/demo"
	"github.com/stretchr/testify/assert"
)

func TestDemo(t *testing.T) {
	// Create a new instance of our service
	service := demo.NewService()

	// Create a new server which handles the HTTP requests for our service
	server := httptest.NewServer(&demo.Handler{Service: service})
	defer server.Close()

	// Create a new client to make RPC calls to the service via the HTTP Handler
	c := demo.NewClient(demo.ClientConfig{Endpoint: server.URL})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Test happy path JSON request and response
	{
		req := demo.SayHelloRequest{
			Name: "Admiral Thrawn",
		}
		var resp demo.SayHelloResponse
		assert.NoError(t, c.SayHello(ctx, &req, &resp))
		assert.Equal(t, "Hello, Admiral Thrawn", resp.Message)
	}

	// Test happy path Protobuf request and response
	//{
	//	req := demo.RenderPixelRequest{
	//		Complexity: 1024,
	//		Height:     2048,
	//		Width:      2048,
	//		I:          1,
	//		J:          1,
	//	}
	//
	//	var resp demo.RenderPixelResponse
	//	assert.NoError(t, c.RenderPixel(ctx, &req, &resp))
	//	assert.Equal(t, 0, resp.Gray)
	//
	//}
}

// TODO: Setup Error cases for tests
//if err := c.SayHello(ctx, &req, &resp); err != nil {
//assert.NoError(t, err)
//var de duh.Error
//if errors.As(err, &de) {
//
//	//de.Details()
//	//de.Error()
//	////msg := `HTTP failed on 'GET https://example.com' (X-Account-Id: '', X-Other-Thing: '') with '404' message 'Not Found'`
//	//msg := `GET https://example.com failed with 'Not Found' message 'Fido is not in the Pet Shop'`
//	//msg = `GET https://example.com failed with 'Request Failed' message 'while reading response body: EOF'`
//	//msg = `GET https://example.com failed with 'Request Failed' message 'while parsing response body: expected {'`
//	//fmt.Printf(msg)
//
//}

// TODO: Start Testing the client!

// Is this a retryable error?
// Is this an infra error?
// Is this a failure?
// Can I tell the diff between an infra error and an error from the service?
