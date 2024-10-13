FROM oven/bun:1 as base
LABEL com.zeabur.image-type="containerized"

ARG entry

WORKDIR /src
COPY package.json bun.lockb* ./
RUN bun install
COPY . .
ENTRYPOINT [ "bun", "run", entry ]
