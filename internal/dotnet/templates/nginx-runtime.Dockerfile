{{define "nginx-runtime"}}
FROM nginx:alpine AS runtime
ENV PORT=8080
WORKDIR /usr/share/nginx/html
COPY --from=build /app/wwwroot ./static/
RUN echo "{{ template "nginx-conf" . }}" > ./nginx.conf
CMD ["/bin/sh" , "-c" , "envsubst '$PORT' < /usr/share/nginx/html/nginx.conf | tee /etc/nginx/conf.d/default.conf && exec nginx -g 'daemon off;'"]
{{end}}
