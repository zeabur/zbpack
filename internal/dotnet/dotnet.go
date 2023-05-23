// Package dotnet is the planner of Dotnet projects.
package dotnet

import (
	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

// GenerateDockerfile generates the Dockerfile for Dotnet projects.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	sdkVer := meta["sdk"]
	entryPoint := meta["entryPoint"]
	dockerfile := `
		# https://hub.docker.com/_/microsoft-dotnet
		FROM mcr.microsoft.com/dotnet/sdk:` + sdkVer + ` AS build
		WORKDIR /source
		
		# copy csproj and restore as distinct layers
		COPY *.csproj ./
		RUN dotnet restore
		
		# copy everything else and build app
		COPY . ./
		WORKDIR /source
		RUN dotnet publish -c release -o /app
		
		# final stage/image
		FROM mcr.microsoft.com/dotnet/aspnet:` + sdkVer + `
		WORKDIR /app
		COPY --from=build /app ./
		ENTRYPOINT ["dotnet", "` + entryPoint + `.dll"]	
	`

	return dockerfile, nil
}

type pack struct {
	*identify
}

// NewPacker returns a new Dotnet packer.
func NewPacker() packer.Packer {
	return &pack{
		identify: &identify{},
	}
}

func (p *pack) GenerateDockerfile(meta types.PlanMeta) (string, error) {
	return GenerateDockerfile(meta)
}

var _ packer.Packer = (*pack)(nil)
