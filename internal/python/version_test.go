package python

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPython3Version(t *testing.T) {
	version := getNodeVersion("^3.8")
	assert.Equal(t, version, "3.8")
}

func TestGetPython3Version_WithMaxVersion(t *testing.T) {
	version := getNodeVersion(">=3.8")
	assert.Equal(t, version, "3.8")
}

func TestGetPython3Version_WithGreaterVersion(t *testing.T) {
	version := getNodeVersion(">3.8")
	assert.Equal(t, version, "3.9")
}
