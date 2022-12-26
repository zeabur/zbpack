package static

import "github.com/zeabur/zbpack/internal/types"

func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	dockerfile := `FROM nginx:alpine
WORKDIR /static
COPY . . 
RUN echo "server { listen 8080; root /static; location / {try_files \$uri /index.html; }}"> /etc/nginx/conf.d/default.conf
EXPOSE 8080`

	return dockerfile, nil
}
