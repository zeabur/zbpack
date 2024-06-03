# https://hub.docker.com/_/microsoft-dotnet
FROM mcr.microsoft.com/dotnet/sdk:{{.DotnetVer}} AS build
WORKDIR /source

# copy csproj and restore as distinct layers
# it works only without a submodule
{{ if .SubmoduleDir | eq "" }}
COPY *.csproj ./
RUN dotnet restore
{{ end }}

# copy everything else and build app
COPY . ./
WORKDIR /source/{{.SubmoduleDir}}
RUN dotnet publish -c release -o /app

# final stage/image
{{ if .Static }}{{ template "nginx-runtime" . }}{{ else }}
FROM mcr.microsoft.com/dotnet/aspnet:{{.DotnetVer}}
ENV PORT=8080
WORKDIR /app
COPY --from=build /app ./
CMD ASPNETCORE_URLS=http://+:$PORT dotnet {{.Out}}.dll
{{ end }}
