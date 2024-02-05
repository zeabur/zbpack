FROM docker.io/lukemathwalker/cargo-chef:latest-rust-1 AS chef
WORKDIR /src

# use sparse to speed up the dependencies download process
ENV CARGO_REGISTRIES_CRATES_IO_PROTOCOL=sparse

# use lld as the linker
RUN apt update \
  && apt install -y lld

RUN mkdir /.cargo && \
  printf '[build]\nrustflags = ["-C", "link-arg=-fuse-ld=lld"]\n' > /.cargo/config.toml

FROM chef AS planner
COPY . .
RUN cargo chef prepare --recipe-path recipe.json

FROM chef AS builder
COPY --from=planner /src/recipe.json recipe.json

# Build dependencies - this is the caching Docker layer!
RUN cargo chef cook --release --recipe-path recipe.json

# Build the project to get the executable file
COPY . .
RUN cargo build --release

# Copy the exe and files listed in .zeabur-preserve to /app/bin
RUN mkdir -p /app/bin \
  # move the files to preserve to /app
  && (cat .zeabur-preserve | xargs -I {} cp -r {} /app/{}) \
  # move the binary to the root of the container
  && (cp target/release/* /app/bin || true)

# {{if not (eq .serverless "true")}}
FROM docker.io/library/debian:bookworm-slim AS runtime

# {{if eq .NeedOpenssl "yes"}}
RUN apt-get update \
  && apt-get install -y openssl \
  && rm -rf /var/lib/apt/lists/*
{{ end }}

RUN useradd -m -s /bin/bash zeabur
COPY --from=builder --chown=zeabur:zeabur /app /app

USER zeabur
WORKDIR /app

ENV BINDIR="/app/bin"
ENV BINNAME="{{ .BinName }}"
ENV EXEFILE="${BINDIR}/${BINNAME}"

RUN if [ ! -x "${EXEFILE}" ]; then \
    find . -type f -executable -print | head -n 1 > EXEFILE; \
  else \
    echo "${EXEFILE}" > EXEFILE; \
  fi

CMD "$(cat EXEFILE)"
# {{ else }}
FROM scratch
COPY --from=builder /app/bin/{{ .BinName }} main
# {{ end }}
