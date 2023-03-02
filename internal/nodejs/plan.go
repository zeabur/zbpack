package nodejs

import (
	"encoding/json"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	. "github.com/zeabur/zbpack/pkg/types"
)

func DeterminePackageManager(absPath string) NodePackageManager {
	if _, err := os.Stat(path.Join(absPath, "yarn.lock")); err == nil {
		return NodePackageManagerYarn
	}

	if _, err := os.Stat(path.Join(absPath, "pnpm-lock.yaml")); err == nil {
		return NodePackageManagerPnpm
	}

	if _, err := os.Stat(path.Join(absPath, "package-lock.json")); err == nil {
		return NodePackageManagerNpm
	}

	return NodePackageManagerYarn
}

func DetermineProjectFramework(absPath string) NodeProjectFramework {

	packageJsonMarshal, err := os.ReadFile(path.Join(absPath, "package.json"))
	if err != nil {
		return NodeProjectFrameworkNone
	}

	packageJson := struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}{}

	if err := json.Unmarshal(packageJsonMarshal, &packageJson); err != nil {
		return NodeProjectFrameworkNone
	}

	if _, isSvelte := packageJson.DevDependencies["svelte"]; isSvelte {
		return NodeProjectFrameworkSvelte
	}

	if _, isHexo := packageJson.Dependencies["hexo"]; isHexo {
		return NodeProjectFrameworkHexo
	}

	if _, isQwik := packageJson.DevDependencies["@builder.io/qwik"]; isQwik {
		return NodeProjectFrameworkQwik
	}

	if _, isVitepress := packageJson.DevDependencies["vitepress"]; isVitepress {
		return NodeProjectFrameworkVitepress
	}

	if _, isVite := packageJson.DevDependencies["vite"]; isVite {
		return NodeProjectFrameworkVite
	}

	if _, isUmi := packageJson.Dependencies["umi"]; isUmi {
		return NodeProjectFrameworkUmi
	}

	if _, isNextJs := packageJson.Dependencies["next"]; isNextJs {
		return NodeProjectFrameworkNextJs
	}

	if _, isNestJs := packageJson.Dependencies["@nestjs/core"]; isNestJs {
		return NodeProjectFrameworkNestJs
	}

	if _, isRemix := packageJson.Dependencies["@remix-run/react"]; isRemix {
		return NodeProjectFrameworkRemix
	}

	if _, isCreateReactApp := packageJson.Dependencies["react-scripts"]; isCreateReactApp {
		return NodeProjectFrameworkCreateReactApp
	}

	if _, isNuxtJs := packageJson.Dependencies["nuxt"]; isNuxtJs {
		return NodeProjectFrameworkNuxtJs
	}

	if _, isNuxtJs := packageJson.DevDependencies["nuxt"]; isNuxtJs {
		return NodeProjectFrameworkNuxtJs
	}

	if _, isVueCliApp := packageJson.DevDependencies["@vue/cli-service"]; isVueCliApp {
		return NodeProjectFrameworkVueCli
	}

	return NodeProjectFrameworkNone

}

func DetermineNeedPuppeteer(absPath string) string {
	packageJsonMarshal, err := os.ReadFile(path.Join(absPath, "package.json"))
	if err != nil {
		return "false"
	}

	packageJson := struct {
		Dependencies map[string]string `json:"dependencies"`
	}{}

	if err := json.Unmarshal(packageJsonMarshal, &packageJson); err != nil {
		return "false"
	}

	if _, hasPuppeteer := packageJson.Dependencies["puppeteer"]; hasPuppeteer {
		return "true"
	}

	return "false"
}

func GetBuildCommand(absPath string) string {
	packageJsonMarshal, err := os.ReadFile(path.Join(absPath, "package.json"))
	if err != nil {
		return ""
	}

	packageJson := struct {
		Scripts map[string]string `json:"scripts"`
	}{}

	if err := json.Unmarshal(packageJsonMarshal, &packageJson); err != nil {
		return ""
	}

	if _, ok := packageJson.Scripts["build"]; ok {
		return "build"
	}

	for key := range packageJson.Scripts {
		if strings.Contains(key, "build") {
			return key
		}
	}

	return ""
}

func GetStartCommand(absPath string) string {
	packageJsonMarshal, err := os.ReadFile(path.Join(absPath, "package.json"))
	if err != nil {
		return ""
	}

	packageJson := struct {
		Scripts         map[string]string `json:"scripts"`
		DevDependencies map[string]string `json:"devDependencies"`
	}{}

	if err := json.Unmarshal(packageJsonMarshal, &packageJson); err != nil {
		return ""
	}

	if _, ok := packageJson.DevDependencies["@builder.io/qwik"]; ok {
		if _, ok := packageJson.Scripts["deploy"]; ok {
			return "deploy"
		}
	}

	if _, ok := packageJson.Scripts["start"]; ok {
		return "start"
	}

	return ""
}

func GetNodeVersion(absPath string) string {
	packageJsonMarshal, err := os.ReadFile(path.Join(absPath, "package.json"))
	if err != nil {
		return ""
	}

	packageJson := struct {
		Engines struct {
			Node string `json:"node"`
		} `json:"engines"`
	}{}

	if err := json.Unmarshal(packageJsonMarshal, &packageJson); err != nil {
		return "16"
	}

	if packageJson.Engines.Node == "" {
		return "16"
	}

	// for example, ">=16.0.0 <17.0.0"
	versionRange := packageJson.Engines.Node

	isVersion, _ := regexp.MatchString(`^\d+(\.\d+){0,2}$`, versionRange)
	if isVersion {
		return versionRange
	}

	// from given version range, we want to extract the minimum version
	// for example, "16"
	ranges := strings.Split(versionRange, " ")
	minVer := -1
	equalMin := false
	maxVer := -1
	equalMax := false
	for _, r := range ranges {
		if strings.HasPrefix(r, ">=") {
			minVerString := strings.TrimPrefix(r, ">=")
			minVerMajor := strings.Split(minVerString, ".")[0]
			minVer, _ = strconv.Atoi(minVerMajor)
			equalMin = true
		} else if strings.HasPrefix(r, ">") {
			minVerString := strings.TrimPrefix(r, ">")
			minVerMajor := strings.Split(minVerString, ".")[0]
			minVer, _ = strconv.Atoi(minVerMajor)
			equalMin = false
		} else if strings.HasPrefix(r, "<=") {
			maxVerString := strings.TrimPrefix(r, "<=")
			maxVerMajor := strings.Split(maxVerString, ".")[0]
			maxVer, _ = strconv.Atoi(maxVerMajor)
			equalMax = true
		} else if strings.HasPrefix(r, "<") {
			maxVerString := strings.TrimPrefix(r, "<")
			maxVerMajor := strings.Split(maxVerString, ".")[0]
			maxVer, _ = strconv.Atoi(maxVerMajor)
			equalMax = false
		}
	}

	if minVer == -1 && maxVer == -1 {
		return "16"
	}

	if minVer == -1 {
		if equalMax {
			if maxVer != 14 && maxVer != 16 && maxVer != 18 {
				return "16"
			}
			return strconv.Itoa(maxVer)
		} else {
			return strconv.Itoa(maxVer - 1)
		}
	}

	if maxVer == -1 {
		if equalMin {
			if minVer != 14 && minVer != 16 && minVer != 18 {
				return "16"
			}
			return strconv.Itoa(minVer)
		} else {
			return strconv.Itoa(minVer + 1)
		}
	}

	return "16"
}

func GetMainFile(absPath string) string {
	packageJsonMarshal, err := os.ReadFile(path.Join(absPath, "package.json"))
	if err != nil {
		return ""
	}

	packageJson := struct {
		Main string `json:"main"`
	}{}

	if err := json.Unmarshal(packageJsonMarshal, &packageJson); err != nil {
		return ""
	}

	return packageJson.Main
}
