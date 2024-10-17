FROM dart:3.2.5
LABEL com.zeabur.image-type="containerized"

ARG build

WORKDIR /app
COPY . .
RUN dart pub get
RUN ${build}
CMD ["/app/bin/main", "--apply-migrations"]
