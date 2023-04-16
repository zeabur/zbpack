package nodejs

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/afero"
)

type PackageJsonEngine struct {
	Node string `json:"node"`
}

// PackageJson is the structure of `package.json`.
type PackageJson struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Scripts         map[string]string `json:"scripts"`
	Engines         PackageJsonEngine `json:"engines"`
	Main            string            `json:"main"`
}

// NewPackageJson returns a new instance of PackageJson
// with some default values.
func NewPackageJson() PackageJson {
	// we don't need to allocate an map for Dependencies,
	// DevDependencies and Scripts, since we won't set them
	// in the null state.
	return PackageJson{}
}

// DeserializePackageJson deserializes a package.json file
// from source. When the deserialize failed, it returns an
// empty PackageJson with the error.
func DeserializePackageJson(fs afero.Fs) (PackageJson, error) {
	p := NewPackageJson()

	packageJsonMarshal, err := afero.ReadFile(fs, "package.json")
	if err != nil {
		return p, fmt.Errorf("read file: %w", err)
	}

	if err := json.Unmarshal(packageJsonMarshal, &p); err != nil {
		return p, fmt.Errorf("unmarshal: %w", err)
	}

	return p, nil
}
