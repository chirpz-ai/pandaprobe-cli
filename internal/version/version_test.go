package version

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	info := Get()
	assert.NotEmpty(t, info.GoVersion)
	assert.NotEmpty(t, info.OS)
	assert.NotEmpty(t, info.Arch)
	assert.Equal(t, Version, info.Version)
}

func TestUserAgent(t *testing.T) {
	assert.True(t, strings.HasPrefix(UserAgent(), "pandaprobe-cli/"))
}
