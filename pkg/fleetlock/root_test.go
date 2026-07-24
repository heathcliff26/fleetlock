package fleetlock

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRootCommand(t *testing.T) {
	cmd := NewRootCommand()

	assert.Equal(t, Name, cmd.Use)
}

func TestExecuteConfigLoadError(t *testing.T) {
	if os.Getenv("RUN_CRASH_TEST") == "1" {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{"fleetlock", "-c", "/nonexistent/config.yaml"}
		Execute()
		os.Exit(0)
	}
	execExitTest(t, "TestExecuteConfigLoadError", true)
}

func TestExecuteK8sClientError(t *testing.T) {
	if os.Getenv("RUN_CRASH_TEST") == "1" {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{"fleetlock", "-c", "testdata/kubernetes-draintimeout.yaml"}
		Execute()
		os.Exit(0)
	}
	execExitTest(t, "TestExecuteK8sClientError", true)
}

func TestExecuteServerCreationError(t *testing.T) {
	if os.Getenv("RUN_CRASH_TEST") == "1" {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{"fleetlock", "-c", "testdata/invalid-storage.yaml"}
		Execute()
		os.Exit(0)
	}
	execExitTest(t, "TestExecuteServerCreationError", true)
}

func TestExecuteServerRunError(t *testing.T) {
	if os.Getenv("RUN_CRASH_TEST") == "1" {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{"fleetlock", "-c", "testdata/ssl-missing-files.yaml"}
		Execute()
		os.Exit(0)
	}
	execExitTest(t, "TestExecuteServerRunError", true)
}

func execExitTest(t *testing.T, test string, exitsError bool) {
	cmd := exec.Command(os.Args[0], "-test.run="+test)
	cmd.Env = append(os.Environ(), "RUN_CRASH_TEST=1")
	err := cmd.Run()
	if exitsError && err == nil {
		t.Fatal("Process exited without error")
	} else if !exitsError && err == nil {
		return
	}
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}
