FROM dart:stable-sdk AS build
ARG build

WORKDIR /app
COPY . .
RUN dart pub get
RUN ${build}

FROM alpine:latest
LABEL com.zeabur.image-type="containerized"

COPY --from=build /app/bin/main /main
CMD ["/main"]
