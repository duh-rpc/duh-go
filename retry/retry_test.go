package retry_test

import (
	"context"
	"fmt"
	"time"

	"github.com/duh-rpc/duh-go/retry"
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

func ExampleOn() {
	c := NewClient()
	var resp DoThingResponse

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// --- Retryable errors are ---
	// 409 - Conflict
	// 429 - Too Many Requests (Will calculate the appropriate interval to retry based on the response)
	// 500 - Internal Error (Hopefully this is a temporary error, and the server will recover)
	// 502,503,504 - Infrastructure errors which hopefully will resolve on retry.

	// `retry.Twice` is the policy that governs when to retry.
	// The `retry.Twice` policy will retry 'twice' if the server responded with one of the following retryable errors
	// The `retry.Twice` uses the default back off policy
	err := retry.On(ctx, retry.Twice, func(ctx context.Context, attempt int) error {
		return c.DoThing(ctx, &DoThingRequest{}, &resp)
	})

	// The `retry.UntilSuccess` policy will retry on retryable errors until success, using the default back off policy.
	err = retry.On(ctx, retry.UntilSuccess, func(ctx context.Context, attempt int) error {
		return c.DoThing(ctx, &DoThingRequest{}, &resp)
	})

	// The `retry.OnConflict` policy will retry only on 409 conflict or until success, using the default back off policy.
	err = retry.On(ctx, retry.OnConflict, func(ctx context.Context, attempt int) error {
		return c.DoThing(ctx, &DoThingRequest{}, &resp)
	})

	customPolicy := retry.Policy{
		Interval: retry.BackOff{
			Min:    time.Millisecond,
			Max:    time.Millisecond * 100,
			Factor: 2,
		},
		OnCodes:  []int64{409, 429, 502, 503, 504},
		Attempts: 5,
	}

	// Users can define a custom retry policy to suit their needs
	err = retry.On(ctx, customPolicy, func(ctx context.Context, attempt int) error {
		return c.DoThing(ctx, &DoThingRequest{}, &resp)
	})

	customPolicy = retry.Policy{
		// No Backoff, just sleep in-between retries
		Interval: retry.Sleep(time.Second),
		OnCodes:  []int64{409, 429, 502, 503, 504},
		// Attempts of 0 indicate infinite retries
		Attempts: 0,
	}

	err = retry.On(ctx, customPolicy, func(ctx context.Context, attempt int) error {
		return c.DoThing(ctx, &DoThingRequest{}, &resp)
	})

	if err != nil {
		fmt.Printf("Error was: %s", err)
	}
}
