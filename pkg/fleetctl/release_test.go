package fleetctl

import (
	"bytes"
	"net/http"
	"os"
	"testing"

	"github.com/heathcliff26/fleetlock/pkg/fake"
	"github.com/stretchr/testify/assert"
)

func TestNewReleaseCommand(t *testing.T) {
	cmd := NewReleaseCommand()

	assert := assert.New(t)

	assert.Equal("release", cmd.Use)
	assert.True(cmd.HasLocalFlags())
}

func TestReleaseCommand(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		srv := fake.NewFakeServer(t, http.StatusOK, "/v1/steady-state")
		defer srv.Close()

		cmd := NewReleaseCommand()
		cmd.SetArgs([]string{srv.URL()})

		b := &bytes.Buffer{}
		cmd.SetOut(b)

		err := cmd.Execute()

		assert.NoError(t, err)
		assert.Contains(t, b.String(), "Success")
	})

	t.Run("MissingArgs", func(t *testing.T) {
		cmd := NewReleaseCommand()

		err := cmd.Execute()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
	})
}

func TestReleaseCommandExitError(t *testing.T) {
	if os.Getenv("RUN_CRASH_TEST") == "1" {
		cmd := NewReleaseCommand()
		cmd.SetArgs([]string{"http://127.0.0.1:1"})
		_ = cmd.Execute()
		os.Exit(0)
	}
	execExitTest(t, "TestReleaseCommandExitError", true)
}
