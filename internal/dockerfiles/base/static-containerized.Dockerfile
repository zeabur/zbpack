FROM caddy AS runtime
ENV PORT=8080

RUN echo -e ":8080 { \n \
    root * /usr/share/caddy \n \
    file_server \n \
}" > /etc/caddy/Caddyfile
WORKDIR /usr/share/caddy
