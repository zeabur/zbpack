package java

import (
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/zeabur/zbpack/internal/utils"

	. "github.com/zeabur/zbpack/pkg/types"
)

func DetermineProjectType(absPath string) JavaProjectType {
	if utils.HasFile(absPath, "pom.xml", "pom.yml", "pom.yaml") {
		return JavaProjectTypeMaven
	}

	if utils.HasFile(absPath, "build.gradle", "build.gradle.kts") {
		return JavaProjectTypeGradle
	}

	return JavaProjectTypeNone
}

func DetermineFramework(pj JavaProjectType, absPath string) JavaFramework {
	if pj == JavaProjectTypeMaven {
		if utils.HasFile(absPath, "pom.xml") {
			pom, err := os.ReadFile(path.Join(absPath, "pom.xml"))
			if err != nil {
				return JavaFrameworkNone
			}

			if strings.Contains(string(pom), "spring-boot-starter-parent") {
				return JavaFrameworkSpringBoot
			}
		}
	}

	if pj == JavaProjectTypeGradle {
		if utils.HasFile(absPath, "build.gradle") {
			gradle, err := os.ReadFile(path.Join(absPath, "build.gradle"))
			if err != nil {
				return JavaFrameworkNone
			}

			if strings.Contains(string(gradle), "org.springframework.boot") {
				return JavaFrameworkSpringBoot
			}
		}
	}

	return JavaFrameworkNone
}

func DetermineJDKVersion(pj JavaProjectType, absPath string) string {

	defaultVersion := "17"

	if pj == JavaProjectTypeMaven {
		if utils.HasFile(absPath, "pom.xml") {
			pom, err := os.ReadFile(path.Join(absPath, "pom.xml"))
			if err != nil {
				return defaultVersion
			}
			r := []string{
				`<java.version>(.*)</java.version>`,
				`<maven.compiler.source>(.*)</maven.compiler.source>`,
				`<maven.compiler.target>(.*)</maven.compiler.target>`,
			}
			for _, v := range r {
				re := regexp.MustCompile(v)
				matches := re.FindStringSubmatch(string(pom))
				if len(matches) > 1 {
					if matches[1] == "1.8" {
						return "8"
					}
					return matches[1]
				}
			}
		}
		return defaultVersion
	}

	if pj == JavaProjectTypeGradle {
		if utils.HasFile(absPath, "build.gradle") {
			gradle, err := os.ReadFile(path.Join(absPath, "build.gradle"))
			if err != nil {
				return defaultVersion
			}
			r := []string{
				`sourceCompatibility = (.*)`,
				`targetCompatibility = (.*)`,
			}
			for _, v := range r {
				re := regexp.MustCompile(v)
				matches := re.FindStringSubmatch(string(gradle))
				if len(matches) > 1 {
					if matches[1] == "1.8" {
						return "8"
					}
					return strings.ReplaceAll(matches[1], "'", "")
				}
			}
		}
		return defaultVersion
	}

	return defaultVersion
}
