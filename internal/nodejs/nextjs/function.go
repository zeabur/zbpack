package nextjs

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	cp "github.com/otiai10/copy"
)

// constructNextFunction will construct the first function page, used as symlinks for other function pages
func constructNextFunction(zeaburOutputDir, tmpDir string) error {
	p := path.Join(zeaburOutputDir, "functions", "__next.func")

	err := os.MkdirAll(p, 0755)
	if err != nil {
		return fmt.Errorf("create function dir: %w", err)
	}

	launcher, err := renderLauncher(tmpDir)
	if err != nil {
		return fmt.Errorf("render launcher: %w", err)
	}

	err = os.WriteFile(path.Join(p, "index.js"), []byte(launcher), 0644)
	if err != nil {
		return fmt.Errorf("write launcher: %w", err)
	}

	err = cp.Copy(path.Join(tmpDir, ".next"), path.Join(p, ".next"))
	if err != nil {
		return fmt.Errorf("copy .next: %w", err)
	}

	err = cp.Copy(path.Join(tmpDir, "package.json"), path.Join(p, "package.json"))
	if err != nil {
		return fmt.Errorf("copy package.json: %w", err)
	}

	outputNodeModulesDir := path.Join(p, "node_modules")
	err = os.MkdirAll(outputNodeModulesDir, 0755)
	if err != nil {
		return fmt.Errorf("create node_modules dir: %w", err)
	}

	var deps []string
	err = filepath.Walk(path.Join(tmpDir, ".next"), func(p string, info os.FileInfo, err error) error {
		if strings.HasSuffix(p, ".nft.json") {
			type nftJSON struct {
				Files []string `json:"files"`
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return fmt.Errorf("read nft.json: %w", err)
			}
			var nft nftJSON
			err = json.Unmarshal(b, &nft)
			if err != nil {
				return fmt.Errorf("unmarshal nft.json: %w", err)
			}
			for _, file := range nft.Files {
				if !strings.Contains(file, "node_modules") {
					continue
				}
				resolved := path.Join(path.Dir(p), file)
				if strings.HasPrefix(resolved, path.Join(tmpDir, "node_modules")) {
					deps = append(deps, resolved)
				}
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk .next: %w", err)
	}

	for _, dep := range deps {
		to := strings.Replace(dep, tmpDir, p, 1)

		err = os.MkdirAll(path.Dir(to), 0755)
		if err != nil {
			return fmt.Errorf("mkdirall: %w", err)
		}

		stat, err := os.Lstat(dep)
		if err != nil {
			return fmt.Errorf("lstat: %w", err)
		}

		if stat.Mode()&os.ModeSymlink == os.ModeSymlink {

			alias, err := os.Readlink(dep)
			if err != nil {
				return fmt.Errorf("read symlink: %w", err)
			}

			err = os.Symlink(alias, to)
			if err != nil && !os.IsExist(err) {
				return fmt.Errorf("write symlink: %w", err)
			}

		} else {

			read, err := os.ReadFile(dep)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}

			err = os.WriteFile(to, read, 0644)
			if err != nil {
				return fmt.Errorf("write file: %w", err)
			}

		}
	}

	return nil
}
