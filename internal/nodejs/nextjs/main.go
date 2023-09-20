// Package nextjs is used to transform build output of Next.js app to the serverless build output format of Zeabur
package nextjs

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	uuid2 "github.com/google/uuid"
	cp "github.com/otiai10/copy"
	"github.com/zeabur/zbpack/internal/utils"
	"github.com/zeabur/zbpack/pkg/types"
)

// TransformServerless will transform build output of Next.js app to the serverless build output format of Zeabur
// It is trying to implement the same logic as build function of https://github.com/vercel/vercel/tree/main/packages/next/src/index.ts
func TransformServerless(image, workdir string) error {

	// create a tmpDir to store the build output of Next.js app
	uuid := uuid2.New().String()
	tmpDir := path.Join(os.TempDir(), uuid)
	defer func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			log.Printf("remove tmp dir: %s\n", err)
		}
	}()

	// /tmpDir/uuid/.next
	nextOutputDir := path.Join(tmpDir, ".next")

	// /tmpDir/uuid/.next/server/pages
	nextOutputServerPagesDir := path.Join(nextOutputDir, "server/pages")

	// /workDir/.zeabur/output
	zeaburOutputDir := path.Join(workdir, ".zeabur/output")

	err := utils.CopyFromImage(image, "/src/.next", tmpDir)
	if err != nil {
		return err
	}

	err = utils.CopyFromImage(image, "/src/node_modules", tmpDir)
	if err != nil {
		return err
	}

	err = utils.CopyFromImage(image, "/src/package.json", tmpDir)
	if err != nil {
		return err
	}

	_ = os.RemoveAll(path.Join(workdir, ".zeabur"))

	var serverlessFunctionPages []string
	internalPages := []string{"_app.js", "_document.js", "_error.js"}
	_ = filepath.Walk(nextOutputServerPagesDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".js") {
			for _, internalPage := range internalPages {
				if strings.HasSuffix(path, internalPage) {
					return nil
				}
			}
			funcPath := strings.TrimPrefix(path, nextOutputServerPagesDir)
			funcPath = strings.TrimSuffix(funcPath, ".js")
			serverlessFunctionPages = append(serverlessFunctionPages, funcPath)
		}
		return nil
	})

	var staticPages []string
	_ = filepath.Walk(nextOutputServerPagesDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".html") {
			filePath := strings.TrimPrefix(path, nextOutputServerPagesDir)
			staticPages = append(staticPages, filePath)
		}
		return nil
	})

	err = os.MkdirAll(path.Join(zeaburOutputDir, "static"), 0755)
	if err != nil {
		return fmt.Errorf("create static dir: %w", err)
	}

	err = cp.Copy(path.Join(nextOutputDir, "static"), path.Join(zeaburOutputDir, "static/_next/static"))
	if err != nil {
		return fmt.Errorf("copy static dir: %w", err)
	}

	err = cp.Copy(path.Join(workdir, "public"), path.Join(zeaburOutputDir, "static"))
	if err != nil {
		return fmt.Errorf("copy public dir: %w", err)
	}

	nextConfig, err := getNextConfig(tmpDir)
	if err != nil {
		return fmt.Errorf("get next config: %w", err)
	}

	tmpl, err := template.New("launcher").Parse(launcherTemplate)
	if err != nil {
		return fmt.Errorf("parse launcher template: %w", err)
	}

	type renderLauncherTemplateContext struct {
		NextConfig string
	}

	var launcher strings.Builder
	err = tmpl.Execute(&launcher, renderLauncherTemplateContext{NextConfig: nextConfig})
	if err != nil {
		return fmt.Errorf("render launcher template: %w", err)
	}

	// if there is any serverless function page, create the first function page and symlinks for other function pages
	if len(serverlessFunctionPages) > 0 {

		// create the first function page
		err = constructNextFunction(zeaburOutputDir, serverlessFunctionPages[0], tmpDir)
		if err != nil {
			return fmt.Errorf("construct next function: %w", err)
		}

		// create symlinks for other function pages
		for i, p := range serverlessFunctionPages {
			if i == 0 {
				continue
			}

			funcPath := path.Join(zeaburOutputDir, "functions", p+".func")

			err = os.MkdirAll(path.Dir(funcPath), 0755)
			if err != nil {
				return fmt.Errorf("create function dir: %w", err)
			}

			err = os.Symlink(path.Join(zeaburOutputDir, "functions", serverlessFunctionPages[0]+".func"), funcPath)
			if err != nil {
				return fmt.Errorf("create symlink: %w", err)
			}
		}
	}

	// copy static pages which is rendered by Next.js at build time, so they will be served as static files
	for _, p := range staticPages {
		err = cp.Copy(path.Join(nextOutputDir, "server/pages", p), path.Join(zeaburOutputDir, "static", p))
		if err != nil {
			return fmt.Errorf("copy static page: %w", err)
		}
	}

	cfg := types.ZeaburOutputConfig{Containerized: false, Routes: make([]types.ZeaburOutputConfigRoute, 0)}
	cfgBytes, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	err = os.WriteFile(path.Join(zeaburOutputDir, "config.json"), cfgBytes, 0644)
	if err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
