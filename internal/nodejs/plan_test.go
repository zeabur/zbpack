package nodejs

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
)

func getMockVersionsList() []*semver.Version {
	rawVer := []string{
		"v20",
		"v19",
		"v18",
		"v17",
		"v16",
		"v15",
		"v14",
		"v13",
		"v12",
		"v10",
	}

	versions := make([]*semver.Version, len(rawVer))

	for i, v := range rawVer {
		ver, err := semver.NewVersion(v)
		if err != nil {
			panic(err)
		}

		versions[i] = ver
	}

	return versions
}

func TestGetNodeVersion_Empty(t *testing.T) {
	v := getNodeVersion("", getMockVersionsList())
	assert.Equal(t, defaultNodeVersion, v)
}

func TestGetNodeVersion_Fixed(t *testing.T) {
	v := getNodeVersion("10", getMockVersionsList())
	assert.Equal(t, "10", v)
}

func TestGetNodeVersion_Or(t *testing.T) {
	v := getNodeVersion("^10 || ^12 || ^14", getMockVersionsList())
	assert.Equal(t, "14", v)
}

func TestGetNodeVersion_GreaterThanWithLessThan(t *testing.T) {
	v := getNodeVersion(">=16 <=20", getMockVersionsList())
	assert.Equal(t, "20", v)
}

func TestGetNodeVersion_GreaterThan(t *testing.T) {
	v := getNodeVersion(">=4", getMockVersionsList())
	assert.Equal(t, "20", v)
}

func TestGetNodeVersion_LessThan(t *testing.T) {
	v := getNodeVersion("<18", getMockVersionsList())
	assert.Equal(t, "17", v)
}
