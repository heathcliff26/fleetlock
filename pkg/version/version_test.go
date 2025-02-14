package version

import (
	"runtime"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVersionCommand(t *testing.T) {
	cmd := NewCommand("test")

	assert := assert.New(t)

	assert.Equal("version", cmd.Use)
	assert.NotNil(cmd.PersistentPreRun, "Should have empty function assigned to override parent function")
}

func TestVersion(t *testing.T) {
	assert := assert.New(t)

	buildinfo, _ := debug.ReadBuildInfo()

	assert.Equal(buildinfo.Main.Version, Version(), "Version should return the version from build info")
}

func TestVersionInfoString(t *testing.T) {
	result := VersionInfoString("test")

	lines := strings.Split(result, "\n")

	assert := assert.New(t)

	buildinfo, _ := debug.ReadBuildInfo()

	if !assert.Equal(5, len(lines), "Should have enough lines") {
		t.FailNow()
	}
	assert.Contains(lines[0], "test")
	assert.Contains(lines[1], buildinfo.Main.Version)

	commit := strings.Split(lines[2], ":")
	assert.NotEmpty(strings.TrimSpace(commit[1]))

	assert.Contains(lines[3], runtime.Version())

	assert.Equal("", lines[4], "Should have trailing newline")
}
