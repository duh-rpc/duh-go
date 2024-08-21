package limit_test

import (
	"bytes"
	"errors"
	"github.com/duh-rpc/duh-go"
	"github.com/duh-rpc/duh-go/limit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"strings"
	"testing"
)

func TestNewReader(t *testing.T) {
	// Can read up to the requested limit
	r := limit.NewReader(io.NopCloser(strings.NewReader("twenty regular bytes")), 21)
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.Equal(t, out, []byte("twenty regular bytes"))

	// Returns an error if more bytes than requested are read
	r = limit.NewReader(io.NopCloser(strings.NewReader("more than twenty regular bytes")), 20)
	_, err = io.ReadAll(r)
	require.Error(t, err)
	assert.Equal(t, err.Error(), "exceeds 20B limit")

	// Returned error is of type duh.Error
	var d duh.Error
	assert.True(t, errors.As(err, &d))

	// Returns the correct SI abbreviations
	kb := 1024
	r = limit.NewReader(io.NopCloser(bytes.NewReader(make([]byte, kb+1))), int64(kb))
	_, err = io.ReadAll(r)
	require.Error(t, err)
	assert.Equal(t, err.Error(), "exceeds 1.0kB limit")

	mb := 1024 * 1024
	r = limit.NewReader(io.NopCloser(bytes.NewReader(make([]byte, mb+1))), int64(mb))
	_, err = io.ReadAll(r)
	require.Error(t, err)
	assert.Equal(t, err.Error(), "exceeds 1.0MB limit")
}
