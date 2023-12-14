/*
Copyright 2023 Derrick J Wippler

Licensed under the MIT License, you may obtain a copy of the License at

https://opensource.org/license/mit/ or in the root of this code repo

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package retry_test

import (
	"context"
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
	_ = retry.On(ctx, retry.Twice, func(ctx context.Context, attempt int) error {
		return c.DoThing(ctx, &DoThingRequest{}, &resp)
	})

	// The `retry.UntilSuccess` policy will retry on retryable errors until success, using the default back off policy.
	_ = retry.On(ctx, retry.UntilSuccess, func(ctx context.Context, attempt int) error {
		return c.DoThing(ctx, &DoThingRequest{}, &resp)
	})

	// The `retry.OnConflict` policy will retry only on 409 conflict or until success, using the default back off policy.
	_ = retry.On(ctx, retry.OnConflict, func(ctx context.Context, attempt int) error {
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
	_ = retry.On(ctx, customPolicy, func(ctx context.Context, attempt int) error {
		return c.DoThing(ctx, &DoThingRequest{}, &resp)
	})

	customPolicy = retry.Policy{
		// No Backoff, just sleep in-between retries
		Interval: retry.Sleep(time.Second),
		OnCodes:  []int64{409, 429, 502, 503, 504},
		// Attempts of 0 indicate infinite retries
		Attempts: 0,
	}

	_ = retry.On(ctx, customPolicy, func(ctx context.Context, attempt int) error {
		return c.DoThing(ctx, &DoThingRequest{}, &resp)
	})
}
