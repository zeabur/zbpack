{{define "nginx-runtime"}}
FROM nginx:alpine as runtime

COPY --from=build /src/{{.OutputDir}} /src/.zeabur/output/static
RUN echo "\
    server { \
        listen 8080; \
        root /src/.zeabur/output/static; \
        absolute_redirect off; \
        location / { \
{{ if .SPA }}            try_files \$uri /index.html; \
{{ else }}            try_files \$uri \$uri.html \$uri/index.html /404.html =404; \
{{ end }}        } \
    }"> /etc/nginx/conf.d/default.conf
EXPOSE 8080
{{end}}
