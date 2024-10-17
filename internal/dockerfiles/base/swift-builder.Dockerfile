FROM swift:5.9-jammy as build

RUN export DEBIAN_FRONTEND=noninteractive DEBCONF_NONINTERACTIVE_SEEN=true \
  && apt-get -q update \
  && apt-get -q dist-upgrade -y \
  && apt-get install -y libjemalloc-dev
