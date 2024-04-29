package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/internal/utils"
)

func TestSplitVersion_Empty(t *testing.T) {
	_, err := utils.SplitVersion("")
	assert.ErrorContains(t, err, "empty version")
}

func TestSplitVersion_Major(t *testing.T) {
	parsedVersion, err := utils.SplitVersion("8")
	assert.NoError(t, err)
	assert.Equal(t, utils.Version{
		Major:      8,
		Minor:      0,
		MinorSet:   false,
		Patch:      0,
		PatchSet:   false,
		Prerelease: "",
	}, parsedVersion)
}

func TestSplitVersion_MajorMinor(t *testing.T) {
	parsedVersion, err := utils.SplitVersion("8.1")
	assert.NoError(t, err)
	assert.Equal(t, utils.Version{
		Major:      8,
		Minor:      1,
		MinorSet:   true,
		Patch:      0,
		PatchSet:   false,
		Prerelease: "",
	}, parsedVersion)
}

func TestSplitVersion_MajorMinorPatch(t *testing.T) {
	parsedVersion, err := utils.SplitVersion("8.1.2")
	assert.NoError(t, err)
	assert.Equal(t, utils.Version{
		Major:      8,
		Minor:      1,
		MinorSet:   true,
		Patch:      2,
		PatchSet:   true,
		Prerelease: "",
	}, parsedVersion)
}

func TestSplitVersion_MajorMinorPatchPrerelease(t *testing.T) {
	parsedVersion, err := utils.SplitVersion("8.1.2-beta1")
	assert.NoError(t, err)
	assert.Equal(t, utils.Version{
		Major:      8,
		Minor:      1,
		MinorSet:   true,
		Patch:      2,
		PatchSet:   true,
		Prerelease: "beta1",
	}, parsedVersion)
}

func TestSplitVersion_MajorMinorPatchPrerelease2(t *testing.T) {
	parsedVersion, err := utils.SplitVersion("8.1.2-beta1.1")
	assert.NoError(t, err)
	assert.Equal(t, utils.Version{
		Major:      8,
		Minor:      1,
		MinorSet:   true,
		Patch:      2,
		PatchSet:   true,
		Prerelease: "beta1.1",
	}, parsedVersion)
}

func TestSplitVersion_MajorMinorPatchZero(t *testing.T) {
	parsedVersion, err := utils.SplitVersion("8.1.0")
	assert.NoError(t, err)
	assert.Equal(t, utils.Version{
		Major:      8,
		Minor:      1,
		MinorSet:   true,
		Patch:      0,
		PatchSet:   true,
		Prerelease: "",
	}, parsedVersion)
}

func TestSplitVersion_MajorMinorZero(t *testing.T) {
	parsedVersion, err := utils.SplitVersion("8.0")
	assert.NoError(t, err)
	assert.Equal(t, utils.Version{
		Major:      8,
		Minor:      0,
		MinorSet:   true,
		Patch:      0,
		PatchSet:   false,
		Prerelease: "",
	}, parsedVersion)
}

func TestSplitVersion_MajorZero(t *testing.T) {
	parsedVersion, err := utils.SplitVersion("8.0.1")
	assert.NoError(t, err)
	assert.Equal(t, utils.Version{
		Major:      8,
		Minor:      0,
		MinorSet:   true,
		Patch:      1,
		PatchSet:   true,
		Prerelease: "",
	}, parsedVersion)
}

func TestSplitVersion_MajorMinorWildcard(t *testing.T) {
	parsedVersion, err := utils.SplitVersion("8.*")
	assert.NoError(t, err)
	assert.Equal(t, utils.Version{
		Major:      8,
		Minor:      0,
		MinorSet:   false,
		Patch:      0,
		PatchSet:   false,
		Prerelease: "",
	}, parsedVersion)
}

func TestSplitVersion_MajorMinorPatchWildcard(t *testing.T) {
	parsedVersion, err := utils.SplitVersion("8.1.*")
	assert.NoError(t, err)
	assert.Equal(t, utils.Version{
		Major:      8,
		Minor:      1,
		MinorSet:   true,
		Patch:      0,
		PatchSet:   false,
		Prerelease: "",
	}, parsedVersion)
}

func TestSplitVersion_MajorInvalid(t *testing.T) {
	_, err := utils.SplitVersion("a")
	assert.NotNil(t, err)
}

func TestSplitVersion_MajorMinorInvalid(t *testing.T) {
	_, err := utils.SplitVersion("1.b")
	assert.NotNil(t, err)
}

func TestSplitVersion_MajorMinorPatchInvalid(t *testing.T) {
	_, err := utils.SplitVersion("1.2.c")
	assert.NotNil(t, err)
}

func TestConstraintToVersion_Equal(t *testing.T) {
	v := utils.ConstraintToVersion("=8.0.2", "7")
	assert.Equal(t, "8.0", v)
}

func TestConstraintToVersion_Tilde(t *testing.T) {
	v := utils.ConstraintToVersion("~8.0.1", "7")
	assert.Equal(t, "8.0", v)
}

func TestConstraintToVersion_Caret(t *testing.T) {
	v := utils.ConstraintToVersion("^8.0", "7")
	assert.Equal(t, "8", v)
}

func TestConstraintToVersion_GreaterThan(t *testing.T) {
	v := utils.ConstraintToVersion(">8.0", "7")
	assert.Equal(t, "8", v)
}

func TestConstraintToVersion_GreaterThanOrEqual(t *testing.T) {
	v := utils.ConstraintToVersion(">=8.0", "7")
	assert.Equal(t, "8", v)
}

func TestConstraintToVersion_LessThan(t *testing.T) {
	v := utils.ConstraintToVersion("<8.0", "7")
	assert.Equal(t, "7", v)
}

func TestConstraintToVersion_LessThanB(t *testing.T) {
	v := utils.ConstraintToVersion("<8", "7")
	assert.Equal(t, "7", v)
}

func TestConstraintToVersion_LessThanOrEqual(t *testing.T) {
	v := utils.ConstraintToVersion("<=8.0", "7")
	assert.Equal(t, "8", v)
}

func TestConstraintToVersion_Explicit(t *testing.T) {
	v := utils.ConstraintToVersion("8.0.2", "7")
	assert.Equal(t, "8.0", v)
}

func TestConstraintToVersion_SecondCaret(t *testing.T) {
	v := utils.ConstraintToVersion("^8.0 ^8.1", "7")
	assert.Equal(t, "8", v)
}

func TestConstraintToVersion_SecondTilde(t *testing.T) {
	v := utils.ConstraintToVersion("~8.0 ~8.1", "7")
	assert.Equal(t, "8.1", v)
}

func TestConstraintToVersion_SecondGreaterThan(t *testing.T) {
	v := utils.ConstraintToVersion("~8.0 >8.1", "7")
	assert.Equal(t, "8", v)
}

func TestConstraintToVersion_SecondOrOperator(t *testing.T) {
	v := utils.ConstraintToVersion("~8.0 || ~8.1", "7")
	assert.Equal(t, "8.1", v)
}

func TestConstraintToVersion_MajorOnly(t *testing.T) {
	v := utils.ConstraintToVersion("8", "7")
	assert.Equal(t, "8", v)
}
