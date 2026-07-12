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

package duh_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/duh-rpc/duh.go/v2"
)

// A HandleBytes handler may render its own error response instead of the standard
// JSON/protobuf Reply — an HTML error page, say. Write the response directly to the
// http.ResponseWriter held in closure scope, then return nil so HandleBytes does not
// also send a Reply. Return a non-nil error only to get the standard Reply — never
// both, or two responses are written.
//
// Writing through the BytesWriter commits a 200 on the first write, so to set a
// different status (404 here) write to the raw ResponseWriter.
func ExampleHandleBytes_renderOwnError() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		duh.HandleBytes(w, r, func(r *http.Request, bw duh.BytesWriter) error {
			// The requested record was not found; answer with a custom HTML page.
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("<html><body>not found</body></html>"))
			return nil // nil == "I answered; don't send a Reply"
		})
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		panic(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)

	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Header.Get("Content-Type"))
	fmt.Println(string(body))
	// Output:
	// 404
	// text/html; charset=utf-8
	// <html><body>not found</body></html>
}
