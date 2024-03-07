package python

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPython3Version(t *testing.T) {
	version := getPython3Version("^3.8")
	assert.Equal(t, version, "3")
}

func TestGetPython3Version_WithMaxVersion(t *testing.T) {
	version := getPython3Version(">=3.8")
	assert.Equal(t, version, "3")
}

func TestGetPython3Version_WithGreaterVersion(t *testing.T) {
	version := getPython3Version(">3.8")
	assert.Equal(t, version, "3")
}

func TestGetPython3Version_WithNullVersion(t *testing.T) {
	version := getPython3Version("")
	assert.Equal(t, version, defaultPython3Version)
}

func TestGetPython3Version_WithInvalidVersion_2(t *testing.T) {
	version := getPython3Version(">w<")
	assert.Equal(t, version, defaultPython3Version)
}

func TestGetPython3Version_WithInvalidVersion_4(t *testing.T) {
	version := getPython3Version("^=====^")
	assert.Equal(t, version, defaultPython3Version)
}
