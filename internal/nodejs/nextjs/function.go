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
func constructNextFunction(zeaburOutputDir, firstFuncPage, tmpDir string) error {
	p := path.Join(zeaburOutputDir, "functions", firstFuncPage+".func")
	if firstFuncPage == "/" {
		p = path.Join(zeaburOutputDir, "functions", "index.func")
	}

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
	err = filepath.Walk(path.Join(tmpDir, ".next"), func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".nft.json") {
			type nftJSON struct {
				Files []string `json:"files"`
			}
			b, err := os.ReadFile(path)
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
				file = file[strings.Index(file, "node_modules"):]
				deps = append(deps, file)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk .next: %w", err)
	}

	for _, dep := range deps {
		err = cp.Copy(path.Join(tmpDir, dep), path.Join(p, dep))
		if err != nil {
			return fmt.Errorf("copy dep: %w", err)
		}
	}

	return nil
}
