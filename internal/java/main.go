package java

import "github.com/zeabur/zbpack/pkg/types"

func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	projectType := meta["type"]
	framework := meta["framework"]

	isMaven := projectType == string(types.JavaProjectTypeMaven)
	isGradle := projectType == string(types.JavaProjectTypeGradle)
	isSpringBoot := framework == string(types.JavaFrameworkSpringBoot)

	// TODO: select the correct base image jdk version
	baseImage := "openjdk:8-jdk-alpine"
	if isMaven {
		baseImage = "maven:latest"
	}
	if isGradle {
		baseImage = "gradle:latest"
	}

	dockerfile := ""

	switch projectType {
	case string(types.JavaProjectTypeMaven):
		dockerfile += `FROM ` + baseImage + ` 
WORKDIR /src
COPY . .
RUN mvn clean dependency:list install
`
	case string(types.JavaProjectTypeGradle):
		dockerfile += `FROM ` + baseImage + `
WORKDIR /src
COPY . .
RUN gradle build
`
	}

	startCmd := ""

	if isMaven {
		startCmd = "CMD java -jar target/*.jar"
	}

	if isGradle {
		startCmd = "CMD java -jar build/libs/*.jar"
	}

	if isMaven && isSpringBoot {
		startCmd = "CMD java -Dserver.port=$PORT -jar target/*.jar"
	}

	if isGradle && isSpringBoot {
		startCmd = "CMD java -Dserver.port=$PORT -jar build/libs/*.jar"
	}

	dockerfile += startCmd

	return dockerfile, nil
}
