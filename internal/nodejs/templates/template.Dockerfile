{{ if .Bun -}}
# Install bun if we need it
FROM oven/bun:1.0 as bun-runtime
{{- end }}

FROM node:{{.NodeVersion}} as build

ENV PORT=8080
WORKDIR /src

# Install additional runtime
{{ if .Bun -}}
# Copy the bun binary from the bun-runtime stage directly.
# A bit hacky but it works.
COPY --from=bun-runtime /usr/local/bin/bun /usr/local/bin
COPY --from=bun-runtime /usr/local/bin/bunx /usr/local/bin
{{- end }}

RUN corepack enable && corepack prepare --all
COPY . .

RUN {{ .InstallCmd }}

# Build if we can build it
{{ if .BuildCmd }}RUN {{ .BuildCmd }}{{ end }}

{{ if .OutputDir }}{{ template "nginx-runtime" . }}{{ else }}
EXPOSE 8080
CMD {{ .StartCmd }}
{{ end }}
