FROM docker.io/denoland/deno AS base
ARG entry

WORKDIR /app
COPY . .
EXPOSE 8080
RUN deno cache ${entry}

FROM base AS run-basic
LABEL com.zeabur.image-type="containerized"

CMD ["run", "--allow-net", "--allow-env", "--allow-read", "--allow-write", "--allow-run", entry]

FROM base AS run-task
LABEL com.zeabur.image-type="containerized"

CMD ["task", "start"]
