FROM node:{{.NodeVersion}}

ENV PORT=8080
WORKDIR /src

COPY package.json package-lock.json* yarn.lock* pnpm-lock.yaml* ./

RUN {{ .InstallCmd }}

COPY . .

# Build if we can build it
{{ if .BuildCmd }}RUN {{ .BuildCmd }}{{ end }}

{{ if .OutputDir }}{{ template "nginx-runtime" . }}{{ else }}
EXPOSE 8080
CMD {{ .StartCmd }}
{{ end }}
