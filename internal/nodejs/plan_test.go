package nodejs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNodeVersion_Empty(t *testing.T) {
	v := getNodeVersion("")
	assert.Equal(t, defaultNodeVersion, v)
}

func TestGetNodeVersion_Fixed(t *testing.T) {
	v := getNodeVersion("10")
	assert.Equal(t, "10", v)
}

func TestGetNodeVersion_Or(t *testing.T) {
	v := getNodeVersion("^10 || ^12 || ^14")
	assert.Equal(t, "14", v)
}

func TestGetNodeVersion_GreaterThanWithLessThan(t *testing.T) {
	v := getNodeVersion(">=16 <=20")
	assert.Equal(t, "20", v)
}

func TestGetNodeVersion_GreaterThan(t *testing.T) {
	v := getNodeVersion(">=4")
	assert.Equal(t, "20", v)
}

func TestGetNodeVersion_LessThan(t *testing.T) {
	v := getNodeVersion("<18")
	assert.Equal(t, "17", v)
}

func TestGetNodeVersion_Exact(t *testing.T) {
	v := getNodeVersion("16.0.0")
	assert.Equal(t, "16.0.0", v)
}

func TestGetNodeVersion_Exact_WithEqualOp(t *testing.T) {
	v := getNodeVersion("=16.0.0")
	assert.Equal(t, "16.0.0", v)
}

func TestGetNodeVersion_CaretMinor(t *testing.T) {
	v := getNodeVersion("^16.1.0")
	assert.Equal(t, "16", v)
}

func TestGetNodeVersion_TildeMinor(t *testing.T) {
	v := getNodeVersion("~16.0.1")
	assert.Equal(t, "16.0", v)
}

func TestGetNodeVersion_ExactWithWildcard(t *testing.T) {
	v := getNodeVersion("16.0.*")
	assert.Equal(t, "16.0", v)
}

func TestGetNodeVersion_TildeWithWildcard(t *testing.T) {
	v := getNodeVersion("~16.*")
	assert.Equal(t, "16", v)
}
