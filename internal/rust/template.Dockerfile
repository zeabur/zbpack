FROM docker.io/library/rust:bookworm AS builder

WORKDIR /src

# use sparse to speed up the dependencies download process
ENV CARGO_REGISTRIES_CRATES_IO_PROTOCOL=sparse

RUN apt update \
  && apt install mold

COPY . .

# Build the project to get the executable file
RUN mold -run cargo build --release

RUN mkdir -p /app/bin \
  # move the files to preserve to /app
  && (cat .zeabur-preserve | xargs -I {} cp -r {} /app/{}) \
  # move the binary to the root of the container
  && (cp target/release/* /app/bin || true)

FROM docker.io/library/debian:bookworm-slim

# {{if eq .NeedOpenssl "yes"}}
RUN apt-get update \
  && apt-get install -y openssl \
  && rm -rf /var/lib/apt/lists/*
# {{ end }}

ENV BINDIR="/app/bin"
ENV BINNAME="{{ .BinName }}"
ENV EXEFILE="${BINDIR}/${BINNAME}"

RUN useradd -m -s /bin/bash zeabur
COPY --from=builder --chown=zeabur:zeabur /app /app

USER zeabur
WORKDIR /app

RUN if [ ! -x "${EXEFILE}" ]; then \
    find . -type f -executable -print | head -n 1 > EXEFILE; \
  else \
    echo "${EXEFILE}" > EXEFILE; \
  fi


CMD "$(cat EXEFILE)"
