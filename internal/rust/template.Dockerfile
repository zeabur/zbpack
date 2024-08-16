FROM rust:1 AS builder

WORKDIR /app
COPY . /app

{{ if ne .BuildCommand "" }}
RUN {{ .BuildCommand }}
{{ end }}

# output to /out/bin
RUN mkdir /out && cargo install --path "{{ .AppDir }}" --root /out

FROM rust:1 AS post-builder

COPY --from=builder /out/bin /app

{{ range .Assets }}
COPY --from=builder /app/{{ . }} /app/{{ . }}
{{ end }}

WORKDIR /app

# Rename the entry point to /app/main
RUN if [ -x "{{ .Entry }}" ]; then \
	mv "{{ .Entry }}" /app/main; \
  else \
  	real_endpoint="$(find . -type f -executable -print | head -n 1)" \
		&& mv "${real_endpoint}" /app/main; \
  fi


{{ if .Serverless }}
FROM scratch
COPY --from=post-builder /app .
{{ else }}
FROM rust:1-slim AS runtime

{{ if .OpenSSL }}
RUN apt-get update \
  && apt-get install -y openssl \
  && rm -rf /var/lib/apt/lists/*
{{ end }}

{{ if ne .PreStartCommand "" }}
RUN {{ .PreStartCommand }}
{{ end }}

COPY --from=post-builder /app /app
{{ if ne .StartCommand "" }}
CMD {{ .StartCommand }}
{{ else }}
CMD ["/app/main"]
{{ end }}

{{ end }}
