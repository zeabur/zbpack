package java

import "github.com/zeabur/zbpack/internal/types"

func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	projectType := meta["type"]

	dockerfile := ""

	switch projectType {
	case string(types.JavaProjectTypeMaven):
		dockerfile += `FROM maven:3.6.3-openjdk-15 as builder
WORKDIR /src
COPY . .
RUN mvn install
`
	case string(types.JavaProjectTypeGradle):
		dockerfile += `FROM gradle:6.8.3-jdk15 as builder
WORKDIR /src
COPY . .
RUN gradle build
`
	}

	dockerfile += `FROM openjdk:15 as runtime
WORKDIR /app
COPY --from=builder /src/target/*.jar .
CMD java -jar *.jar
`

	return dockerfile, nil
}
