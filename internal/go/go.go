package _go

import "github.com/zeabur/zbpack/pkg/types"

func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	return `FROM docker.io/library/golang:1.18 as builder
RUN mkdir /src
WORKDIR /src
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . /src/
ENV CGO_ENABLED=0
RUN go build -o ./bin/server ` + meta["entry"] + `

FROM alpine as runtime
COPY --from=builder /src/bin/server /bin/server
ENV PORT=8080
EXPOSE 8080
CMD ["/bin/server"]
`, nil
}
