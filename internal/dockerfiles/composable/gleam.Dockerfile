FROM ghcr.io/gleam-lang/gleam:v1.3.2-erlang-alpine AS build
RUN apk add --no-cache elixir
RUN mix local.hex --force
RUN mix local.rebar --force
COPY . /build/
RUN cd /build \
  && gleam export erlang-shipment \
  && mv build/erlang-shipment /app \
  && rm -r /build

FROM build AS runtime-containerized
LABEL com.zeabur.image-type="containerized"
WORKDIR /app
ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["run"]

FROM scratch AS runtime-serverless
LABEL com.zeabur.image-type="serverless"
LABEL com.zeabur.serverless-transformer="gleam"
COPY --from=build /app /
