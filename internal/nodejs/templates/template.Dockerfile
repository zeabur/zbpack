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

RUN corepack enable

{{ .InstallCmd }}

COPY . .
{{ if and (eq .Framework "nuxt.js") .Serverless }}
ENV NITRO_PRESET=node
{{ end }}
# Build if we can build it
{{ if .BuildCmd }}RUN {{ .BuildCmd }}{{ end }}
{{ if .Serverless }}
FROM scratch as output
COPY --from=build /src /
{{ else if ne .OutputDir "" }}
FROM scratch as output
COPY --from=build /src/{{ .OutputDir }} /
{{ else }}
EXPOSE 8080
CMD {{ .StartCmd }}{{ end }}
