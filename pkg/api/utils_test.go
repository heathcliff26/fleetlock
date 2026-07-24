package api

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseResponse(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		res, err := ParseResponse(io.NopCloser(strings.NewReader(`{"kind":"ok","value":"test"}`)))

		assert.NoError(t, err)
		assert.Equal(t, "ok", res.Kind)
		assert.Equal(t, "test", res.Value)
	})

	t.Run("Invalid", func(t *testing.T) {
		_, err := ParseResponse(io.NopCloser(strings.NewReader("not-json")))

		assert.Error(t, err)
	})
}
