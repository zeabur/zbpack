package java

import (
	"github.com/zeabur/zbpack/internal/source"
	"regexp"
	"strings"

	"github.com/zeabur/zbpack/internal/utils"

	. "github.com/zeabur/zbpack/pkg/types"
)

func DetermineProjectType(src *source.Source) JavaProjectType {
	if utils.HasFile(src, "pom.xml", "pom.yml", "pom.yaml") {
		return JavaProjectTypeMaven
	}

	if utils.HasFile(src, "build.gradle", "build.gradle.kts") {
		return JavaProjectTypeGradle
	}

	return JavaProjectTypeNone
}

func DetermineFramework(pj JavaProjectType, src *source.Source) JavaFramework {
	if pj == JavaProjectTypeMaven {
		if utils.HasFile(src, "pom.xml") {
			pom, err := (*src).ReadFile("pom.xml")
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
			gradle, err := (*src).ReadFile("build.gradle")
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

func DetermineJDKVersion(pj JavaProjectType, src *source.Source) string {

	defaultVersion := "17"

	if pj == JavaProjectTypeMaven {
		if utils.HasFile(src, "pom.xml") {
			pom, err := (*src).ReadFile("pom.xml")
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
			gradle, err := (*src).ReadFile("build.gradle")
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
