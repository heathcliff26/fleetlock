package fleetctl

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRootCommand(t *testing.T) {
	cmd := NewRootCommand()

	assert.Equal(t, Name, cmd.Use)
}

func TestExecute(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"fleetctl", "--help"}

	Execute()
}

func TestExecuteError(t *testing.T) {
	if os.Getenv("RUN_CRASH_TEST") == "1" {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{"fleetctl", "invalid-command"}
		Execute()
		os.Exit(0)
	}
	execExitTest(t, "TestExecuteError", true)
}
