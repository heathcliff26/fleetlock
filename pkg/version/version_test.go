package version

import (
	"runtime"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVersionCommand(t *testing.T) {
	cmd := NewCommand("test")

	assert := assert.New(t)

	assert.Equal("version", cmd.Use)
	assert.NotNil(cmd.PersistentPreRun, "Should have empty function assigned to override parent function")
}

func TestVersion(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(gitVersion, Version(), "Version should return the version from gitVersion")
}

func TestVersionInfoString(t *testing.T) {
	oldGitCommit := gitCommit
	defer func() { gitCommit = oldGitCommit }()

	gitCommit = "1234567890abcdef"

	result := VersionInfoString("test")

	lines := strings.Split(result, "\n")

	assert := assert.New(t)

	buildinfo, _ := debug.ReadBuildInfo()

	require.Len(t, lines, 5, "Should have enough lines")
	assert.Contains(lines[0], "test")
	assert.Contains(lines[1], buildinfo.Main.Version)

	commit := strings.Split(lines[2], ":")
	assert.Equal("1234567", strings.TrimSpace(commit[1]), "commit hash should be truncated")

	assert.Contains(lines[3], runtime.Version())

	assert.Equal("", lines[4], "Should have trailing newline")
}

func TestInitGitCommit(t *testing.T) {
	oldGitCommit := gitCommit
	defer func() { gitCommit = oldGitCommit }()
	assert := assert.New(t)

	gitCommit = "1234567890abcdef"
	initGitCommit()
	assert.Equal("1234567890abcdef", gitCommit, "gitCommit should not be changed")

	gitCommit = "$Format:%H$"
	initGitCommit()
	assert.NotEqual("$Format:%H$", gitCommit, "gitCommit should be changed")

	gitCommit = ""
	initGitCommit()
	assert.Equal("Unknown", gitCommit, "gitCommit should be Unknown")
}
