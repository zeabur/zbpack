{{- if .Bun -}}
# Install bun if we need it
FROM oven/bun:1.0 as bun-runtime
{{ end -}}
FROM node:{{.NodeVersion}} as build

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

{{ if eq .ServiceDir "" }}COPY . .{{ end }}
{{ if and (eq .Framework "nuxt.js") .Serverless }}
ENV NITRO_PRESET=node
{{ end }}
{{ if and (eq .Framework "nuxt.js") (not .Serverless) }}
ENV NITRO_PRESET=node-server
{{ end }}
# Build if we can build it
{{ if .BuildCmd }}RUN {{ .BuildCmd }}{{ end }}
{{ if .Serverless }}
FROM scratch as output
COPY --from=build /src/{{ .ServiceDir }} /
{{ else if ne .OutputDir "" }}
FROM scratch as output
COPY --from=build /src/{{ .ServiceDir }}/{{ .OutputDir }} /
{{ else }}
EXPOSE 8080
CMD {{ .StartCmd }}{{ end }}
