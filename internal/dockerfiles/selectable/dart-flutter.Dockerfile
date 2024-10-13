FROM zeabur/zbpack-dart-flutter-base AS build
ARG build

WORKDIR /app
COPY . .
RUN flutter clean
RUN flutter pub get
RUN ${build}

FROM scratch AS target-static
LABEL com.zeabur.image-type="static"

COPY --from=build /app/build/web /

FROM docker.io/library/caddy AS target-containerized
LABEL com.zeabur.image-type="containerized"

COPY --from=build /app/build/web /usr/share/caddy
