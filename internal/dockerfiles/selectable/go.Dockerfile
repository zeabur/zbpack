# Build stage
FROM zeabur/zbpack-go-builder AS builder

# Argument for CGO setting
ARG cgo
ARG build
ARG entry

ENV CGO_ENABLED=${cgo}

RUN mkdir /src
WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . /src/

RUN if [ -n "$build" ]; then $build; fi

RUN go build -o ./bin/server $entry

FROM scratch AS target-serverless
LABEL com.zeabur.image-type="serverless"
LABEL com.zeabur.serverless-transformer="golang"
COPY --from=builder /src/bin/server main

FROM alpine as target-containerized
LABEL com.zeabur.image-type="containerized"
COPY --from=builder /src/bin/server /bin/server
CMD ["/bin/server"]
