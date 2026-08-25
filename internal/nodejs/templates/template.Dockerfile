FROM node:{{.NodeVersion}} AS build

ENV PORT=8080
WORKDIR /src

{{ .InitCmd }}
COPY . .
{{ .InstallCmd }}

# Build if we can build it
{{ if .BuildCmd }}RUN {{ .BuildCmd }}{{ if isNitro .Framework }}
RUN if [ -d .zeabur/output ]; then echo "zbpack: Nitro preset 'zeabur' emits unsupported serverless output (.zeabur/output). Use NITRO_PRESET=node-server (or bun) instead of NITRO_PRESET=zeabur / SERVER_PRESET=zeabur." >&2; exit 1; fi{{ end }}{{ end }}
{{ if ne .OutputDir "" }}
FROM scratch AS output
COPY --from=build /src/{{ .AppDir }}/{{ .OutputDir }} /
FROM zeabur/caddy-static AS runtime
COPY --from=output / /usr/share/caddy
{{ else }}
EXPOSE 8080
CMD {{ .StartCmd }}{{ end }}
