FROM node:{{.NodeVersion}} as build

ENV PORT=8080
WORKDIR /src

RUN corepack enable && corepack prepare --all

# Install dependencies and create a cache layer.
COPY package.json* package-lock.json* yarn.lock* pnpm-lock.yaml* .npmrc* ./
RUN {{ .InstallCmd }}

COPY . .

# Try to install again in case there were something we didn't catch
RUN {{ .InstallCmd }}

# Build if we can build it
{{ if .BuildCmd }}RUN {{ .BuildCmd }}{{ end }}

{{ if .OutputDir }}{{ template "nginx-runtime" . }}{{ else }}
EXPOSE 8080
CMD {{ .StartCmd }}
{{ end }}
