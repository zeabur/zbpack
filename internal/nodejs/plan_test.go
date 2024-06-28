package nodejs

import (
	"strconv"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/pkg/plan"
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
	assert.Equal(t, "4", v) // FIXME: should be the latest?
}

func TestGetNodeVersion_LessThan(t *testing.T) {
	v := getNodeVersion("<18")
	assert.Equal(t, "17", v)
}

func TestGetNodeVersion_Exact(t *testing.T) {
	v := getNodeVersion("16.0.0")
	assert.Equal(t, "16.0", v)
}

func TestGetNodeVersion_Exact_WithEqualOp(t *testing.T) {
	v := getNodeVersion("=16.0.0")
	assert.Equal(t, "16.0", v)
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

func TestGetNodeVersion_NvmRcLts(t *testing.T) {
	v := getNodeVersion("lts/*")
	assert.Equal(t, strconv.FormatUint(maxLtsNodeVersion, 10), v)
}

func TestGetNodeVersion_NvmRcLatest(t *testing.T) {
	v := getNodeVersion("node")
	assert.Equal(t, strconv.FormatUint(maxNodeVersion, 10), v)
}

func TestGetNodeVersion_VPrefixedVersion(t *testing.T) {
	v := getNodeVersion("v20.11.0")
	assert.Equal(t, "20.11", v)
}

func TestGetInstallCmd_CustomizeInstallCmd(t *testing.T) {
	src := afero.NewMemMapFs()
	_ = afero.WriteFile(src, "package.json", []byte(`{}`), 0o644)

	config := plan.NewProjectConfigurationFromFs(src, "")
	config.Set(plan.ConfigInstallCommand, "echo 'installed'")

	packageJSON, err := DeserializePackageJSON(src)
	assert.NoError(t, err)

	ctx := &nodePlanContext{
		ProjectPackageJSON: packageJSON,
		Config:             config,
		Src:                src,
	}
	installlCmd := GetInstallCmd(ctx)

	// RUN should be provided in planMeta
	assert.Contains(t, installlCmd, "RUN ")

	// for customized installation command, no cache are allowed.
	assert.Contains(t, installlCmd, "COPY . .")

	// the installation command should be contained
	assert.Contains(t, installlCmd, "echo 'installed'")
}

func TestGetInstallCmd_DefaultInstallCmd(t *testing.T) {
	src := afero.NewMemMapFs()
	_ = afero.WriteFile(src, "package.json", []byte(`{}`), 0o644)
	_ = afero.WriteFile(src, "yarn.lock", []byte(``), 0o644)

	config := plan.NewProjectConfigurationFromFs(src, "")

	packageJSON, err := DeserializePackageJSON(src)
	assert.NoError(t, err)

	ctx := &nodePlanContext{
		ProjectPackageJSON: packageJSON,
		Config:             config,
		Src:                src,
	}

	installlCmd := GetInstallCmd(ctx)

	// RUN should be provided in planMeta
	assert.Contains(t, installlCmd, "RUN ")

	// for default installation command, cache are allowed.
	assert.Contains(t, installlCmd, "COPY yarn.lock* .")

	// the installation command should be contained
	assert.Contains(t, installlCmd, "yarn install")
}

func TestGetInstallCmd_CustomizeInstallCmdDeps(t *testing.T) {
	src := afero.NewMemMapFs()
	_ = afero.WriteFile(src, "package.json", []byte(`{
	"dependencies": {
		"playwright-chromium": "*"
	}
}`), 0o644)

	config := plan.NewProjectConfigurationFromFs(src, "")
	config.Set(plan.ConfigInstallCommand, "echo 'installed'")

	packageJSON, err := DeserializePackageJSON(src)
	assert.NoError(t, err)

	ctx := &nodePlanContext{
		ProjectPackageJSON: packageJSON,
		Config:             config,
		Src:                src,
	}
	installlCmd := GetInstallCmd(ctx)

	// RUN should be provided in planMeta
	assert.Contains(t, installlCmd, "RUN ")

	// the playwright dependencies should be installed
	assert.Contains(t, installlCmd, "libnss3 libatk1.0-0 libatk-bridge2.0-0")

	// the installation command should be contained
	assert.Contains(t, installlCmd, "echo 'installed'")
}
