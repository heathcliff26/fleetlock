package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetZincateAppID(t *testing.T) {
	assert := assert.New(t)

	id, err := GetZincateAppID()
	assert.NoError(err, "Should succeed")
	assert.NotEmpty(id, "Should return id")
}

func TestTrimTrailingSlash(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("https://fleetlock.example.com", TrimTrailingSlash("https://fleetlock.example.com"), "Should not change URL")
	assert.Equal("https://fleetlock.example.com", TrimTrailingSlash("https://fleetlock.example.com/"), "Should remove trailing /")
}
