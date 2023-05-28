{{define "nginx-conf"}} \
server { \
    listen \$PORT; \
    \
    location / { \
        root /usr/share/nginx/html/static; \
        try_files \$uri \$uri/ /index.html =404; \
    } \
} \
{{end}}