package java

import (
	"os"
	"path"
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
