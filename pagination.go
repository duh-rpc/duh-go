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

package duh

import (
	"context"

	"github.com/duh-rpc/duh.go/v2/retry"
)

// Page contains cursor metadata returned by a paginated endpoint.
type Page struct {
	EndCursor   string
	HasNextPage bool
}

// IteratorConfig configures an Iterator. A nil *IteratorConfig uses defaults.
type IteratorConfig struct {
	// RetryPolicy retries failed fetch calls. Nil disables retries.
	RetryPolicy *retry.Policy
}

// Iterator provides cursor-based pagination over any paginated endpoint.
// It is NOT safe for concurrent use.
type Iterator[T any] struct {
	fetch  func(ctx context.Context, cursor string) ([]T, Page, error)
	config IteratorConfig
	cursor string
	err    error
	done   bool
}

// NewIterator creates a new pagination iterator.
// The fetch function wraps the caller's RPC call -- the iterator passes the current cursor,
// the caller constructs the request with their own page size and parameters.
func NewIterator[T any](
	fetch func(ctx context.Context, cursor string) ([]T, Page, error),
	conf *IteratorConfig,
) *Iterator[T] {
	var cfg IteratorConfig
	if conf != nil {
		cfg = *conf
	}
	return &Iterator[T]{
		fetch:  fetch,
		config: cfg,
	}
}

// Next populates the provided slice with the next page of results.
// Returns true if a page was fetched, false when iteration is complete or an error occurred.
func (it *Iterator[T]) Next(ctx context.Context, page *[]T) bool {
	if it.done {
		return false
	}

	var items []T
	var pg Page
	var err error

	if it.config.RetryPolicy != nil {
		err = retry.On(ctx, *it.config.RetryPolicy, func(ctx context.Context, _ int) error {
			var fetchErr error
			items, pg, fetchErr = it.fetch(ctx, it.cursor)
			return fetchErr
		})
	} else {
		items, pg, err = it.fetch(ctx, it.cursor)
	}

	if err != nil {
		it.err = err
		it.done = true
		return false
	}

	it.cursor = pg.EndCursor
	if !pg.HasNextPage {
		it.done = true
	}

	*page = items
	return true
}

// Err returns the error that caused iteration to stop, if any.
func (it *Iterator[T]) Err() error {
	return it.err
}
