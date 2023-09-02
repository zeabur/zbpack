// Package java is the planner of Java projects.
package java

import (
	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

// GenerateDockerfile generates the Dockerfile for Java projects.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	projectType := meta["type"]
	framework := meta["framework"]
	jdkVersion := meta["jdk"]
	targetExt := meta["targetExt"]

	isMaven := projectType == string(types.JavaProjectTypeMaven)
	isGradle := projectType == string(types.JavaProjectTypeGradle)
	isSpringBoot := framework == string(types.JavaFrameworkSpringBoot)

	baseImage := "docker.io/library/openjdk:" + jdkVersion + "-jdk-slim"

	dockerfile := ""

	switch projectType {
	case string(types.JavaProjectTypeMaven):
		dockerfile += `FROM ` + baseImage + `
RUN apt-get update && apt-get install -y maven
RUN apt-get install -y ca-certificates-java
WORKDIR /src
COPY . .
RUN mvn clean dependency:list install -Dmaven.test.skip=true
`
	case string(types.JavaProjectTypeGradle):
		baseImage = "docker.io/library/gradle:8.1.0-jdk" + jdkVersion + "-alpine"
		dockerfile += `FROM ` + baseImage + `
WORKDIR /src
COPY . .
RUN gradle build -x test
`
	}

	startCmd := ""
	wildcardFilename := "*." + targetExt

	if isMaven {
		startCmd = "CMD java -jar target/" + wildcardFilename
	}

	if isGradle {
		startCmd = "CMD java -jar build/libs/" + wildcardFilename
	}

	if isMaven && isSpringBoot {
		startCmd = "CMD java -Dserver.port=$PORT -jar target/" + wildcardFilename
	}

	if isGradle && isSpringBoot {
		startCmd = "CMD java -Dserver.port=$PORT -jar build/libs/" + wildcardFilename
	}

	dockerfile += startCmd

	return dockerfile, nil
}

type pack struct {
	*identify
}

// NewPacker returns a new Java packer.
func NewPacker() packer.Packer {
	return &pack{
		identify: &identify{},
	}
}

func (p *pack) GenerateDockerfile(meta types.PlanMeta) (string, error) {
	return GenerateDockerfile(meta)
}

var _ packer.Packer = (*pack)(nil)
