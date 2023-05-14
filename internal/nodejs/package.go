package nodejs

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/afero"
)

// PackageJSONEngine is the structure of `package.json`'s `engines` field.
type PackageJSONEngine struct {
	Node string `json:"node"`
}

// PackageJSON is the structure of `package.json`.
type PackageJSON struct {
	PackageManager  *string           `json:"packageManager"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Scripts         map[string]string `json:"scripts"`
	Engines         PackageJSONEngine `json:"engines"`
	Main            string            `json:"main"`
}

// NewPackageJSON returns a new instance of PackageJson
// with some default values.
func NewPackageJSON() PackageJSON {
	// we don't need to allocate an map for Dependencies,
	// DevDependencies and Scripts, since we won't set them
	// in the null state.
	return PackageJSON{}
}

// DeserializePackageJSON deserializes a package.json file
// from source. When the deserialize failed, it returns an
// empty PackageJson with the error.
func DeserializePackageJSON(source afero.Fs) (PackageJSON, error) {
	p := NewPackageJSON()

	packageJSONMarshal, err := afero.ReadFile(source, "package.json")
	if err != nil {
		return p, fmt.Errorf("read file: %w", err)
	}

	if err := json.Unmarshal(packageJSONMarshal, &p); err != nil {
		return p, fmt.Errorf("unmarshal: %w", err)
	}

	return p, nil
}
