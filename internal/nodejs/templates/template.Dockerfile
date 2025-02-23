FROM node:{{.NodeVersion}} AS build

ENV PORT=8080
WORKDIR /src

{{ .InitCmd }}
COPY . .
{{ .InstallCmd }}

# Build if we can build it
{{ if .BuildCmd }}RUN {{ .BuildCmd }}{{ end }}
{{ if ne .OutputDir "" }}
FROM scratch AS output
COPY --from=build /src/{{ .AppDir }}/{{ .OutputDir }} /
FROM zeabur/caddy-static AS runtime
COPY --from=output / /usr/share/caddy
{{ else }}
EXPOSE 8080
CMD {{ .StartCmd }}{{ end }}
