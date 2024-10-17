FROM ubuntu:jammy

RUN export DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true \
  && apt-get -q update \
  && apt-get -q dist-upgrade -y \
  && apt-get -q install -y \
  libjemalloc2 \
  ca-certificates \
  tzdata \
  libcurl4 \
  libxml2 \
  && rm -r /var/lib/apt/lists/*

ENV SWIFT_BACKTRACE=enable=yes,sanitize=yes,threads=all,images=all,interactive=no,swift-backtrace=./swift-backtrace-static
