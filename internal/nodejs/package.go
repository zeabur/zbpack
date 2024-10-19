package nodejs

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/utils"
)

// PackageJSONEngine is the structure of `package.json`'s `engines` field.
type PackageJSONEngine struct {
	Node string `json:"node"`
	Bun  string `json:"bun,omitempty"`
}

// PackageJSON is the structure of `package.json`.
type PackageJSON struct {
	PackageManager  *string           `json:"packageManager,omitempty"`
	Dependencies    map[string]string `json:"dependencies,omitempty"`
	DevDependencies map[string]string `json:"devDependencies,omitempty"`
	Scripts         map[string]string `json:"scripts,omitempty"`
	Engines         PackageJSONEngine `json:"engines"`
	Main            string            `json:"main"`
	Module          string            `json:"module"`

	// yarn workspace
	Workspaces []string `json:"workspaces,omitempty"`
}

// NewPackageJSON returns a new instance of PackageJson
// with some default values.
func NewPackageJSON() PackageJSON {
	return PackageJSON{
		Dependencies:    make(map[string]string),
		DevDependencies: make(map[string]string),
		Scripts:         make(map[string]string),
	}
}

// DeserializePackageJSON deserializes a package.json file
// from source. When the deserialization failed, it returns an
// empty PackageJson with the error.
func DeserializePackageJSON(source afero.Fs) (PackageJSON, error) {
	p := NewPackageJSON()

	packageJSONMarshal, err := utils.ReadFileToUTF8(source, "package.json")
	if err != nil {
		return p, fmt.Errorf("read file: %w", err)
	}

	if err := json.Unmarshal(packageJSONMarshal, &p); err != nil {
		return p, fmt.Errorf("unmarshal: %w", err)
	}

	return p, nil
}

// FindDependency checks if the package.json contains
// the given dependency in "dependency" and "devDependencies",
// and returns the version of the dependency.
func (p PackageJSON) FindDependency(name string) (string, bool) {
	d, ok := p.Dependencies[name]
	if ok {
		return d, true
	}

	d, ok = p.DevDependencies[name]
	if ok {
		return d, true
	}

	return "", false
}
