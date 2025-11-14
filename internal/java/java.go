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
	javaArgs := meta["javaArgs"]

	isMaven := projectType == string(types.JavaProjectTypeMaven)
	isGradle := projectType == string(types.JavaProjectTypeGradle)
	isSpringBoot := framework == string(types.JavaFrameworkSpringBoot)

	var dockerfile string
	baseImage := "docker.io/library/eclipse-temurin:" + jdkVersion

	switch projectType {
	case string(types.JavaProjectTypeMaven):
		dockerfile += `FROM ` + baseImage + `
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
  --mount=type=cache,target=/var/lib/apt,sharing=locked \
  apt update \
  	&& apt-get --no-install-recommends install -y \
		maven \
		ca-certificates-java
WORKDIR /src
COPY . .
RUN mvn clean dependency:list install -Dmaven.test.skip=true
`
	case string(types.JavaProjectTypeGradle):
		dockerfile += `FROM ` + baseImage + `
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
  --mount=type=cache,target=/var/lib/apt,sharing=locked \
  apt update \
  	&& apt-get --no-install-recommends install -y \
		ca-certificates-java \
		unzip

# https://github.com/gradle/docker-gradle/blob/master/jdk17-jammy/Dockerfile
ENV GRADLE_HOME=/opt/gradle
ENV GRADLE_VERSION=8.14.3
ARG GRADLE_DOWNLOAD_SHA256=bd71102213493060956ec229d946beee57158dbd89d0e62b91bca0fa2c5f3531
RUN set -o errexit -o nounset \
    && echo "Downloading Gradle" \
    && wget --no-verbose --output-document=gradle.zip "https://services.gradle.org/distributions/gradle-${GRADLE_VERSION}-bin.zip" \
    \
    && echo "Checking Gradle download hash" \
    && echo "${GRADLE_DOWNLOAD_SHA256} *gradle.zip" | sha256sum --check - \
    \
    && echo "Installing Gradle" \
    && unzip gradle.zip \
    && rm gradle.zip \
    && mv "gradle-${GRADLE_VERSION}" "${GRADLE_HOME}/" \
    && ln --symbolic "${GRADLE_HOME}/bin/gradle" /usr/bin/gradle

# Disable Gradle daemon
RUN mkdir -p ~/.gradle && echo "org.gradle.daemon=false" >> ~/.gradle/gradle.properties

WORKDIR /src
COPY . .
RUN gradle build -x test

FROM ` + baseImage + `
WORKDIR /src
COPY --from=0 /src/build /src/build
`
	}

	startCmd := ""
	wildcardFilename := "*." + targetExt

	switch {
	case javaArgs != "":
		startCmd = "CMD java " + javaArgs
	case isMaven && isSpringBoot:
		startCmd = "CMD java -Dserver.port=$PORT -jar target/" + wildcardFilename
	case isGradle && isSpringBoot:
		startCmd = "CMD java -Dserver.port=$PORT -jar build/libs/" + wildcardFilename
	case isMaven:
		startCmd = "CMD java -jar target/" + wildcardFilename
	case isGradle:
		startCmd = "CMD java -jar build/libs/" + wildcardFilename
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
