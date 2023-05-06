package java

import (
	"regexp"
	"strings"

	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/utils"

	. "github.com/zeabur/zbpack/pkg/types"
)

func DetermineProjectType(src afero.Fs) JavaProjectType {
	if utils.HasFile(src, "pom.xml", "pom.yml", "pom.yaml") {
		return JavaProjectTypeMaven
	}

	if utils.HasFile(src, "build.gradle", "build.gradle.kts") {
		return JavaProjectTypeGradle
	}

	return JavaProjectTypeNone
}

func DetermineFramework(pj JavaProjectType, src afero.Fs) JavaFramework {
	if pj == JavaProjectTypeMaven {
		if utils.HasFile(src, "pom.xml") {
			pom, err := afero.ReadFile(src, "pom.xml")
			if err != nil {
				return JavaFrameworkNone
			}

			if strings.Contains(string(pom), "spring-boot-starter-parent") {
				return JavaFrameworkSpringBoot
			}
		}
	}

	if pj == JavaProjectTypeGradle {
		if utils.HasFile(src, "build.gradle") {
			gradle, err := afero.ReadFile(src, "build.gradle")
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

func DetermineJDKVersion(pj JavaProjectType, src afero.Fs) string {
	defaultVersion := "17"

	if pj == JavaProjectTypeMaven {
		if utils.HasFile(src, "pom.xml") {
			pom, err := afero.ReadFile(src, "pom.xml")
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
		if utils.HasFile(src, "build.gradle") {
			gradle, err := afero.ReadFile(src, "build.gradle")
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
