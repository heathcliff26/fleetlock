package fleetctl

import (
	"bytes"
	"net/http"
	"os"
	"testing"

	"github.com/heathcliff26/fleetlock/pkg/fake"
	"github.com/stretchr/testify/assert"
)

func TestNewLockCommand(t *testing.T) {
	cmd := NewLockCommand()

	assert := assert.New(t)

	assert.Equal("lock", cmd.Use)
	assert.True(cmd.HasLocalFlags())
}

func TestLockCommand(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		srv := fake.NewFakeServer(t, http.StatusOK, "/v1/pre-reboot")
		defer srv.Close()

		cmd := NewLockCommand()
		cmd.SetArgs([]string{srv.URL()})

		b := &bytes.Buffer{}
		cmd.SetOut(b)

		err := cmd.Execute()

		assert.NoError(t, err)
		assert.Contains(t, b.String(), "Success")
	})

	t.Run("MissingArgs", func(t *testing.T) {
		cmd := NewLockCommand()

		err := cmd.Execute()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
	})
}

func TestLockCommandExitError(t *testing.T) {
	if os.Getenv("RUN_CRASH_TEST") == "1" {
		cmd := NewLockCommand()
		cmd.SetArgs([]string{"http://127.0.0.1:1"})
		_ = cmd.Execute()
		os.Exit(0)
	}
	execExitTest(t, "TestLockCommandExitError", true)
}
