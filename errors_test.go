package duh_test

import (
	"encoding/json"
	"github.com/harbor-pkg/duh"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrorf(t *testing.T) {
	var err error
	err = duh.Errorf(duh.CodeClientError, "while marshalling request body: %s", json.InvalidUnmarshalError{})
	assert.Equal(t, "", err.Error())
}
