FROM node:{{.NodeVersion}} AS build

ENV PORT=8080
WORKDIR /src

{{ .InitCmd }}
COPY . .
{{ .InstallCmd }}

# Build if we can build it
{{ if .BuildCmd }}RUN {{ .BuildCmd }}{{ end }}
{{ if and .Serverless (eq .OutputDir "") }}
FROM scratch AS output
COPY --from=build /src/{{ .AppDir }} /
{{ else if ne .OutputDir "" }}
FROM scratch AS output
COPY --from=build /src/{{ .AppDir }}/{{ .OutputDir }} /
{{ if not .Serverless }}
FROM zeabur/caddy-static AS runtime
COPY --from=output / /usr/share/caddy
{{ end }}
{{ else }}
EXPOSE 8080
CMD {{ .StartCmd }}{{ end }}
