package fleetctl

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIDCommand(t *testing.T) {
	cmd := NewIDCommand()

	assert := assert.New(t)

	assert.Equal("id", cmd.Name(), "Should have correct name")

	assert.True(cmd.HasLocalFlags(), "Should have local flags")
}

func TestIDCommand(t *testing.T) {
	cmd := NewIDCommand()

	assert := assert.New(t)
	require := require.New(t)

	b := bytes.NewBufferString("")
	cmd.SetOut(b)

	err := cmd.Execute()
	assert.NoError(err, "Command should run without error")

	out, err := io.ReadAll(b)
	require.NoError(err, "Output should be parsed without error")
	assert.NotEmpty(string(out), "Should output an id when using local machine id")

	b = bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--" + flagNameMachineID, "dfd7882acda64c34aca76193c46f5d4e"})

	err = cmd.Execute()
	assert.NoError(err, "Command should run without error")

	out, err = io.ReadAll(b)
	require.NoError(err, "Output should be parsed without error")
	assert.Equal("35ba2101ae3f4d45b96e9c51f461bbff\n", string(out), "Should output the correct app id")
}
