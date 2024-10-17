FROM zeabur/zbpack-swift-builder AS build

WORKDIR /build

COPY ./Package.* ./
RUN swift package resolve --skip-update \
  $([ -f ./Package.resolved ] && echo "--force-resolved-versions" || true)

COPY . .

RUN swift build -c release \
  --static-swift-stdlib \
  -Xlinker -ljemalloc

WORKDIR /staging

RUN cp "$(swift build --package-path /build -c release --show-bin-path)/App" ./

RUN cp "/usr/libexec/swift/linux/swift-backtrace-static" ./

RUN find -L "$(swift build --package-path /build -c release --show-bin-path)/" -regex '.*\.resources$' -exec cp -Ra {} ./ \;

RUN [ -d /build/Public ] && { mv /build/Public ./Public && chmod -R a-w ./Public; } || true
RUN [ -d /build/Resources ] && { mv /build/Resources ./Resources && chmod -R a-w ./Resources; } || true

FROM zeabur/zbpack-swift-runtime AS runtime

WORKDIR /app

COPY --from=build --chown=vapor:vapor /staging /app

ENTRYPOINT ["./App"]
CMD ["serve", "--env", "production", "--hostname", "0.0.0.0", "--port", "8080"]
