package duh_test

import (
	"context"
	"fmt"
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
	var resp demo.SayHelloResponse

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	req := demo.SayHelloRequest{
		Name: "Admiral Thrawn",
	}

	if err := c.SayHello(ctx, &req, &resp); err != nil {
		assert.NoError(t, err)
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
	}
	fmt.Println(resp.Message)
}
