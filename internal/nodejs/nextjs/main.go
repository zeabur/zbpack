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

	mapset "github.com/deckarep/golang-set"
	esbuild "github.com/evanw/esbuild/pkg/api"
	uuid2 "github.com/google/uuid"
	cp "github.com/otiai10/copy"
	"github.com/zeabur/zbpack/pkg/types"
)

// TransformServerless will transform build output of Next.js app to the serverless build output format of Zeabur
// It is trying to implement the same logic as build function of https://github.com/vercel/vercel/tree/main/packages/next/src/index.ts
func TransformServerless(workdir string) error {

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

	// /tmpDir/uuid/.next/server/app
	nextOutputServerAppDir := path.Join(nextOutputDir, "server/app")

	// /workDir/.zeabur/output
	zeaburOutputDir := path.Join(workdir, ".zeabur/output")

	fmt.Println("=> Copying build output from image")

	err := cp.Copy(path.Join(os.TempDir(), "/zbpack/buildkit"), path.Join(tmpDir))
	if err != nil {
		return fmt.Errorf("copy buildkit output to tmp dir: %w", err)
	}

	serverlessFunctionPages := mapset.NewSet()

	fmt.Println("=> Collect serverless function pages")

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
			serverlessFunctionPages.Add(funcPath)
		}
		return nil
	})

	_ = filepath.Walk(nextOutputServerAppDir, func(p string, info os.FileInfo, err error) error {

		// if we found any page.js or route.js inside .next/server/app, this is a serverless function
		if strings.HasSuffix(p, "page.js") || strings.HasSuffix(p, "route.js") {
			funcPath := strings.TrimPrefix(p, nextOutputServerAppDir)
			funcPath = path.Dir(funcPath)
			serverlessFunctionPages.Add(funcPath)
		}

		// if we found any .rsc file inside .next/server/app, this is a serverless function
		if strings.HasSuffix(p, ".rsc") {
			funcPath := strings.TrimPrefix(p, nextOutputServerAppDir)
			serverlessFunctionPages.Add(funcPath)
		}

		return nil
	})

	fmt.Println("=> Copying static asset files")

	err = os.MkdirAll(path.Join(zeaburOutputDir, "static"), 0755)
	if err != nil {
		return fmt.Errorf("create static dir: %w", err)
	}

	err = cp.Copy(path.Join(nextOutputDir, "static"), path.Join(zeaburOutputDir, "static/_next/static"))
	if err != nil {
		return fmt.Errorf("copy static dir: %w", err)
	}

	err = cp.Copy(path.Join(workdir, "public"), path.Join(zeaburOutputDir, "static"))
	if err != nil && !strings.Contains(err.Error(), "no such file or directory") {
		return fmt.Errorf("copy public dir: %w", err)
	}

	fmt.Println("=> Constructing Next.js serverless function")

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

	fmt.Println("=> Creating serverless function symlinks")

	// Create the __next function route
	err = constructNextFunction(zeaburOutputDir, tmpDir)
	if err != nil {
		return fmt.Errorf("construct next function: %w", err)
	}

	cfg := types.ZeaburOutputConfig{
		Routes: []types.ZeaburOutputConfigRoute{
			// redirect all requests not match any static files to __next function
			{Src: "/(.*)", Dest: "/__next"},
		},
	}

	cfgBytes, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	err = os.WriteFile(path.Join(zeaburOutputDir, "config.json"), cfgBytes, 0644)
	if err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	fmt.Println("=> Building edge middleware")

	err = buildMiddleware(tmpDir, zeaburOutputDir)
	if err != nil {
		return fmt.Errorf("build middleware: %w", err)
	}

	return nil
}

func buildMiddleware(workdir, zeaburOutputDir string) error {
	files := []string{"middleware.js", "middleware.ts", "src/middleware.js", "src/middleware.ts"}
	var middlewareFile string
	for _, file := range files {
		if _, err := os.Stat(path.Join(workdir, file)); err == nil {
			middlewareFile = file
			break
		}
	}

	if middlewareFile == "" {
		return nil
	}

	res := esbuild.Build(esbuild.BuildOptions{
		EntryPoints:   []string{path.Join(workdir, middlewareFile)},
		Bundle:        true,
		Platform:      esbuild.PlatformNode,
		Loader:        map[string]esbuild.Loader{".wasm": esbuild.LoaderBinary},
		AbsWorkingDir: workdir,
	})
	if res.Errors != nil && len(res.Errors) > 0 {
		println(res.Errors[0].Text)
		return fmt.Errorf("esbuild run failed")
	}

	wp := path.Join(zeaburOutputDir, "functions/_middleware.func/index.js")
	_ = os.MkdirAll(path.Dir(wp), 0755)
	err := os.WriteFile(wp, res.OutputFiles[0].Contents, 0644)
	if err != nil {
		return fmt.Errorf("write middleware: %w", err)
	}

	return nil
}
