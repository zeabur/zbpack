package static

import "github.com/zeabur/zbpack/pkg/types"

func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	dockerfile := `FROM docker.io/library/nginx:alpine as runtime
WORKDIR /usr/share/nginx/html/static
COPY . . 
RUN echo "server { listen 8080; root /usr/share/nginx/html/static; }"> /etc/nginx/conf.d/default.conf
EXPOSE 8080`

	return dockerfile, nil
}
