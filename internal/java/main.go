package java

import "github.com/zeabur/zbpack/pkg/types"

func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	projectType := meta["type"]
	framework := meta["framework"]
	jdkVersion := meta["jdk"]

	isMaven := projectType == string(types.JavaProjectTypeMaven)
	isGradle := projectType == string(types.JavaProjectTypeGradle)
	isSpringBoot := framework == string(types.JavaFrameworkSpringBoot)

	baseImage := "openjdk:" + jdkVersion + "-jdk-slim"

	dockerfile := ""

	switch projectType {
	case string(types.JavaProjectTypeMaven):
		dockerfile += `FROM ` + baseImage + `
RUN apt-get update && apt-get install -y maven
WORKDIR /src
COPY . .
RUN mvn clean dependency:list install
`
	case string(types.JavaProjectTypeGradle):
		dockerfile += `FROM ` + baseImage + `
RUN apt-get update && apt-get install -y gradle
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
