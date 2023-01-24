package php

import (
	_ "embed"
	"github.com/zeabur/zbpack/pkg/types"
	"strings"
)

//go:embed nginx.conf
var nginxConf string

func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	phpVersion := meta["phpVersion"]
	getPhpImage := "FROM php:" + phpVersion + "-fpm\n"

	installCMD := `
RUN apt-get update 
RUN apt-get install -y nginx zip
RUN curl -sS https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer
ADD https://github.com/mlocati/docker-php-extension-installer/releases/latest/download/install-php-extensions /usr/local/bin/
RUN chmod +x /usr/local/bin/install-php-extensions && sync
RUN install-php-extensions zip
`

	nginxConf = strings.ReplaceAll(nginxConf, "\n", "\\n")
	nginxConf = strings.ReplaceAll(nginxConf, "$", "\\$")

	// copy source code to /var/www/public, which is Nginx root directory
	copyCommand := `
COPY --chown=www-data:www-data . /var/www/public
WORKDIR /var/www/public
`

	switch meta["framework"] {
	case "laravel":
		// if laravel, copy source code to /var/www, because laravel has its own public directory
		copyCommand = `
COPY --chown=www-data:www-data . /var/www  
WORKDIR /var/www
`
	}

	// generate Nginx config to let it pass the request to php-fpm
	copyCommand += `
RUN rm /etc/nginx/sites-enabled/default
RUN echo "` + nginxConf + `" >> /etc/nginx/sites-enabled/default
`

	// install dependencies with composer
	composerInstallCmd := `
RUN composer install --optimize-autoloader --no-dev
`

	startCmd := `
EXPOSE ${PORT}
CMD nginx; php-fpm
`

	dockerFile := getPhpImage +
		installCMD +
		copyCommand +
		composerInstallCmd +
		startCmd

	return dockerFile, nil
}
