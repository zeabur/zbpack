{{- if .Bun -}}
# Install bun if we need it
FROM oven/bun:{{.BunVersion}} AS bun-runtime
{{ end -}}
FROM node:{{.NodeVersion}} AS build

ENV PORT=8080
WORKDIR /src

{{- if .Bun }}
# Copy the bun binary from the bun-runtime stage directly.
# A bit hacky but it works.
COPY --from=bun-runtime /usr/local/bin/bun /usr/local/bin
COPY --from=bun-runtime /usr/local/bin/bunx /usr/local/bin
{{- end }}

# Check if we have 'corepack' available; if none, we
# install corepack@0.10.0.
RUN which corepack || npm install -g --force corepack@0.10.0
RUN corepack enable

{{ .InstallCmd }}

{{ if eq .AppDir "" }}COPY . .{{ end }}
{{ if .Framework | isNitro }}
{{ if .Serverless }}
ENV NITRO_PRESET=node
{{ else if and (not .Serverless) (prefixed .StartCmd "bun") }}
ENV NITRO_PRESET=bun
ENV HOST=0.0.0.0
{{ else }}
ENV NITRO_PRESET=node-server
ENV HOST=0.0.0.0
{{ end }}
{{ end }}
# Build if we can build it
{{ if .BuildCmd }}RUN {{ .BuildCmd }}{{ end }}
{{ if .Serverless }}
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
