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

	"github.com/deckarep/golang-set"
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

	// /tmpDir/uuid/.next/server/app
	nextOutputServerAppDir := path.Join(nextOutputDir, "server/app")

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

	serverlessFunctionPages := mapset.NewSet()
	prerenderPaths := mapset.NewSet()
	staticPages := mapset.NewSet()

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

	_ = filepath.Walk(nextOutputServerPagesDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".html") {
			filePath := strings.TrimPrefix(path, nextOutputServerPagesDir)
			staticPages.Add(filePath)
		}
		return nil
	})

	_ = filepath.Walk(nextOutputServerAppDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".html") {
			filePath := strings.TrimPrefix(path, nextOutputServerAppDir)
			staticPages.Add(filePath)
		}
		return nil
	})

	serverlessFunctionPages.Add("/_next/image")

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

	file, err := os.ReadFile(path.Join(nextOutputDir, "prerender-manifest.json"))
	if err != nil {
		return fmt.Errorf("read prerender manifest: %w", err)
	}

	type prerenderManifestRoute struct {
		SrcRoute                 *string         `json:"srcRoute"`
		DataRoute                string          `json:"dataRoute"`
		InitialRevalidateSeconds utils.IntOrBool `json:"initialRevalidateSeconds"`
	}

	type prerenderManifest struct {
		Routes        map[string]prerenderManifestRoute `json:"routes"`
		DynamicRoutes map[string]prerenderManifestRoute `json:"dynamicRoutes"`
	}

	var pm prerenderManifest
	err = json.Unmarshal(file, &pm)
	if err != nil {
		return fmt.Errorf("unmarshal prerender manifest: %w", err)
	}

	for route, config := range pm.Routes {
		if config.InitialRevalidateSeconds.IsInt {
			serverlessFunctionPages.Add(route)
			serverlessFunctionPages.Add(config.DataRoute)
		}
	}

	for _, config := range pm.DynamicRoutes {
		serverlessFunctionPages.Add(config.DataRoute)
		prerenderPaths.Add(config.DataRoute)
	}

	// if there is any serverless function page, create the first function page and symlinks for other function pages
	if serverlessFunctionPages.Cardinality() > 0 {

		// create the first function page
		firstFuncPage := serverlessFunctionPages.Pop().(string)
		err = constructNextFunction(zeaburOutputDir, firstFuncPage, tmpDir)
		if err != nil {
			return fmt.Errorf("construct next function: %w", err)
		}

		// create symlinks for other function pages
		for i := range serverlessFunctionPages.Iter() {
			p := i.(string)
			funcPath := path.Join(zeaburOutputDir, "functions", p+".func")
			if p == "/" {
				funcPath = path.Join(zeaburOutputDir, "functions", "index.func")
			}

			err = os.MkdirAll(path.Dir(funcPath), 0755)
			if err != nil {
				return fmt.Errorf("create function dir: %w", err)
			}

			target := path.Join(zeaburOutputDir, "functions", firstFuncPage+".func")
			if firstFuncPage == "/" {
				target = path.Join(zeaburOutputDir, "functions", "index.func")
			}

			err = os.Symlink(target, funcPath)
			if err != nil && !os.IsExist(err) {
				return fmt.Errorf("create symlink: %w", err)
			}
		}
	}

	for route, config := range pm.Routes {
		if config.InitialRevalidateSeconds.IsInt {
			r := route
			if config.SrcRoute != nil {
				r = *config.SrcRoute
			}
			prerenderPaths.Add(r)
			prerenderPaths.Add(config.DataRoute)
		}
	}

	for r := range prerenderPaths.Iter() {
		err = writePrerenderConfig(zeaburOutputDir, r.(string))
		if err != nil {
			return fmt.Errorf("write prerender config: %w", err)
		}
	}

	// copy static pages which is rendered by Next.js at build time, so they will be served as static files
	for i := range staticPages.Iter() {
		p := i.(string)
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

func writePrerenderConfig(zeaburOutputDir, r string) error {
	prerenderConfigFilename := r + ".prerender-config.json"
	if r == "/" {
		prerenderConfigFilename = "index.prerender-config.json"
	}

	pcPath := path.Join(zeaburOutputDir, "functions", prerenderConfigFilename)

	err := os.MkdirAll(path.Dir(pcPath), 0755)
	if err != nil {
		return fmt.Errorf("create prerender config dir: %w", err)
	}

	err = os.WriteFile(pcPath, []byte("{\"type\": \"Prerender\"}"), 0644)
	if err != nil {
		return fmt.Errorf("write prerender config: %w", err)
	}

	return nil
}
