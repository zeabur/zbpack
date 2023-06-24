// Package php is the planner for PHP projects.
package php

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

//go:embed nginx.conf
var nginxConf string

// GenerateDockerfile generates the Dockerfile for PHP projects.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	phpVersion := meta["phpVersion"]
	getPhpImage := "FROM docker.io/library/php:" + phpVersion + "-fpm\n"

	installCMD := fmt.Sprintf(`
RUN apt-get update
RUN apt-get install -y %s
RUN curl -sS https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer
ADD https://github.com/mlocati/docker-php-extension-installer/releases/latest/download/install-php-extensions /usr/local/bin/
RUN chmod +x /usr/local/bin/install-php-extensions && sync
`, meta["deps"])

	nginxConf = strings.ReplaceAll(nginxConf, "\n", "\\n")
	nginxConf = strings.ReplaceAll(nginxConf, "$", "\\$")

	// copy source code to /var/www/public, which is Nginx root directory
	copyCommand := `
COPY --chown=www-data:www-data . /var/www/public
WORKDIR /var/www/public
`

	if meta["framework"] != "none" {
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
RUN  echo '#!/bin/sh\n\
extensions=$(cat composer.json | jq -r ".require | to_entries[] | select(.key | startswith(\"ext-\")) | .key[4:]")\n\
for ext in $extensions; do\n\
    echo "Installing PHP extension: $ext"\n\
    docker-php-ext-install $ext\n\
done' > /usr/local/bin/install_php_extensions.sh \
    && chmod +x /usr/local/bin/install_php_extensions.sh \
    && /usr/local/bin/install_php_extensions.sh
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

type pack struct {
	*identify
}

// NewPacker returns a new PHP packer.
func NewPacker() packer.Packer {
	return &pack{
		identify: &identify{},
	}
}

func (p *pack) GenerateDockerfile(meta types.PlanMeta) (string, error) {
	return GenerateDockerfile(meta)
}

var _ packer.Packer = (*pack)(nil)
