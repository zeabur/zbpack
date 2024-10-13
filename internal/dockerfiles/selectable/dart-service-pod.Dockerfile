FROM dart:3.2.5 AS build
ARG build

WORKDIR /app
COPY . .
RUN dart pub get
RUN ${build}
CMD ["/app/bin/main", "--apply-migrations"]
