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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/duh-rpc/duh.go/v2"
	v1 "github.com/duh-rpc/duh.go/v2/proto/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestHandleBytesStreamsBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		duh.HandleBytes(w, r, func(r *http.Request, bw duh.BytesWriter) error {
			_, err := bw.Write([]byte("hello, bytes"))
			return err
		})
	}))
	defer server.Close()

	resp, err := http.Post(server.URL, "", nil)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "hello, bytes", string(body))
	assert.Equal(t, duh.ContentOctetStream, resp.Header.Get("Content-Type"))
	assert.Equal(t, duh.DUHVersion, resp.Header.Get(duh.HeaderDUHVersion))
	assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
}

func TestHandleBytesEmptySuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		duh.HandleBytes(w, r, func(r *http.Request, bw duh.BytesWriter) error {
			// The handler succeeds without writing any bytes (e.g. an empty export).
			return nil
		})
	}))
	defer server.Close()

	resp, err := http.Post(server.URL, "", nil)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// A successful empty response still carries the standard service headers.
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Empty(t, body)
	assert.Equal(t, duh.ContentOctetStream, resp.Header.Get("Content-Type"))
	assert.Equal(t, duh.DUHVersion, resp.Header.Get(duh.HeaderDUHVersion))
	assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
}

func TestHandleBytesHandlerOverridesHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		duh.HandleBytes(w, r, func(r *http.Request, bw duh.BytesWriter) error {
			// Override the seeded Content-Type and set an app header before writing.
			bw.Header().Set("Content-Type", "text/plain; charset=utf-8")
			bw.Header().Set("X-RPC-Job-Id", "job-42")
			_, err := bw.Write([]byte("output"))
			return err
		})
	}))
	defer server.Close()

	resp, err := http.Post(server.URL, "", nil)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, "output", string(body))
	assert.Equal(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))
	assert.Equal(t, "job-42", resp.Header.Get("X-RPC-Job-Id"))
}

func TestHandleBytesErrorBeforeFirstByte(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		duh.HandleBytes(w, r, func(r *http.Request, bw duh.BytesWriter) error {
			// The handler fails before producing any output, so HandleBytes can still
			// send a standard error Reply.
			return duh.NewServiceError(duh.CodeBadRequest, "bad input", nil, nil)
		})
	}))
	defer server.Close()

	resp, err := http.Post(server.URL, "", nil)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, duh.DUHVersion, resp.Header.Get(duh.HeaderDUHVersion))
	assert.NotEqual(t, duh.ContentOctetStream, resp.Header.Get("Content-Type"))
	assert.Contains(t, string(body), "bad input")
}

// A content endpoint's pre-body error must fall back to a JSON Reply when the
// client's Accept is the endpoint's own (non-Reply-capable) success media type,
// rather than masking the real status with a 455.
func TestHandleBytesPreBodyErrorFallsBackToJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		duh.HandleBytes(w, r, func(r *http.Request, bw duh.BytesWriter) error {
			return duh.NewServiceError(duh.CodeNotFound, "path not found", nil, nil)
		})
	}))
	defer server.Close()

	for _, test := range []struct {
		name   string
		accept string
	}{
		{name: "octet-stream", accept: duh.ContentOctetStream},
		{name: "text/html", accept: "text/html"},
		{name: "text/html with charset", accept: "text/html; charset=utf-8"},
		{name: "image/png", accept: "image/png"},
	} {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, server.URL, nil)
			require.NoError(t, err)
			req.Header.Set("Accept", test.accept)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// The real 404 survives — no 455 masking it.
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
			assert.Equal(t, duh.ContentTypeJSON, resp.Header.Get("Content-Type"))
			assert.Equal(t, duh.DUHVersion, resp.Header.Get(duh.HeaderDUHVersion))

			var reply v1.Reply
			require.NoError(t, protojson.Unmarshal(body, &reply))
			assert.Equal(t, "404", reply.GetCode())
			assert.Equal(t, "path not found", reply.GetMessage())
		})
	}
}

// A protobuf client of a content endpoint (request Content-Type: application/protobuf)
// gets a protobuf-encoded error Reply, not a force-JSON one.
func TestHandleBytesPreBodyErrorPrefersProtobuf(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		duh.HandleBytes(w, r, func(r *http.Request, bw duh.BytesWriter) error {
			return duh.NewServiceError(duh.CodeUnauthorized, "not allowed", nil, nil)
		})
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodPost, server.URL, nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", duh.ContentTypeProtoBuf)
	req.Header.Set("Accept", duh.ContentOctetStream)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, duh.ContentTypeProtoBuf, resp.Header.Get("Content-Type"))

	var reply v1.Reply
	require.NoError(t, proto.Unmarshal(body, &reply))
	assert.Equal(t, "401", reply.GetCode())
	assert.Equal(t, "not allowed", reply.GetMessage())
}

// A Reply-capable Accept passes through unchanged: an explicit JSON Accept still
// negotiates to JSON even when the request Content-Type is protobuf.
func TestHandleBytesPreBodyErrorRespectsReplyCapableAccept(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		duh.HandleBytes(w, r, func(r *http.Request, bw duh.BytesWriter) error {
			return duh.NewServiceError(duh.CodeBadRequest, "bad input", nil, nil)
		})
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodPost, server.URL, nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", duh.ContentTypeProtoBuf)
	req.Header.Set("Accept", duh.ContentTypeJSON)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, duh.ContentTypeJSON, resp.Header.Get("Content-Type"))

	var reply v1.Reply
	require.NoError(t, protojson.Unmarshal(body, &reply))
	assert.Equal(t, "400", reply.GetCode())
}

// Negotiating the error encoding must not mutate the caller's request: middleware
// that reads r.Header after HandleBytes returns must still see the client's Accept.
func TestHandleBytesPreBodyErrorDoesNotMutateRequest(t *testing.T) {
	observed := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		duh.HandleBytes(w, r, func(r *http.Request, bw duh.BytesWriter) error {
			return duh.NewServiceError(duh.CodeNotFound, "path not found", nil, nil)
		})
		// Outer middleware observes the request after the content handler returns.
		observed <- r.Header.Get("Accept")
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodPost, server.URL, nil)
	require.NoError(t, err)
	req.Header.Set("Accept", "image/png")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// The error still negotiated to a JSON Reply...
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, duh.ContentTypeJSON, resp.Header.Get("Content-Type"))
	// ...but the client's original Accept survives on the request.
	assert.Equal(t, "image/png", <-observed)
}

// A service may override ReplyContentError to produce a custom, non-JSON error
// representation (e.g. an HTML error page) without patching the framework.
func TestHandleBytesReplyContentErrorOverridable(t *testing.T) {
	original := duh.ReplyContentError
	defer func() { duh.ReplyContentError = original }()

	duh.ReplyContentError = func(w http.ResponseWriter, r *http.Request, err error) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("<html><body>not found</body></html>"))
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		duh.HandleBytes(w, r, func(r *http.Request, bw duh.BytesWriter) error {
			return duh.NewServiceError(duh.CodeNotFound, "path not found", nil, nil)
		})
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodPost, server.URL, nil)
	require.NoError(t, err)
	req.Header.Set("Accept", duh.ContentOctetStream)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))
	assert.Contains(t, string(body), "not found")
}

func TestHandleBytesErrorAfterFirstByteIsCommitted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		duh.HandleBytes(w, r, func(r *http.Request, bw duh.BytesWriter) error {
			_, _ = bw.Write([]byte("partial"))
			return duh.NewServiceError(duh.CodeInternalError, "failed mid-stream", nil, nil)
		})
	}))
	defer server.Close()

	resp, err := http.Post(server.URL, "", nil)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Once bytes were written the 200 is committed; the error cannot become a Reply.
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, duh.ContentOctetStream, resp.Header.Get("Content-Type"))
	assert.Equal(t, "partial", string(body))
	assert.NotContains(t, string(body), "failed mid-stream")
}
