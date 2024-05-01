// Package php is the planner for PHP projects.
package php

import (
	"fmt"

	"github.com/zeabur/zbpack/pkg/packer"
	"github.com/zeabur/zbpack/pkg/types"
)

// GenerateDockerfile generates the Dockerfile for PHP projects.
func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	phpVersion := meta["phpVersion"]
	projectProperty := PropertyFromString(meta["property"])
	serverMode := "fpm"

	getPhpImage := "FROM docker.io/library/php:" + phpVersion + "-fpm\n"

	// Custom server for Laravel Octane
	switch meta["octaneServer"] {
	case OctaneServerSwoole:
		getPhpImage = "FROM docker.io/phpswoole/swoole:php" + phpVersion + "\n" +
			"RUN docker-php-ext-install pcntl\n"
		serverMode = "swoole"
	}

	environmentInstallCmd := "\n"
	if meta["deps"] != "" {
		environmentInstallCmd += fmt.Sprintf(`RUN apt-get update && apt-get install -y %s && rm -rf /var/lib/apt/lists/*
`, meta["deps"])
	}
	if projectProperty&types.PHPPropertyComposer != 0 {
		environmentInstallCmd += "RUN curl -sS https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer\n"
	}
	if meta["exts"] != "" {
		environmentInstallCmd += `ADD https://github.com/mlocati/docker-php-extension-installer/releases/latest/download/install-php-extensions /usr/local/bin/
RUN chmod +x /usr/local/bin/install-php-extensions && sync
RUN install-php-extensions ` + meta["exts"] + "\n"
	}

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

	if serverMode == "fpm" {
		// generate Nginx config to let it pass the request to php-fpm
		nginxConf, err := RetrieveNginxConf(meta["app"])
		if err != nil {
			return "", fmt.Errorf("retrieve nginx conf: %w", err)
		}

		copyCommand += `
RUN rm /etc/nginx/sites-enabled/default
RUN echo "` + nginxConf + `" >> /etc/nginx/sites-enabled/default
`
	}

	// install project dependencies
	projectInstallCmd := "\n"
	if projectProperty&types.PHPPropertyComposer != 0 {
		projectInstallCmd += `RUN composer install --optimize-autoloader --no-dev` + "\n"
	}

	startCmd := "\n" + "CMD " + meta["startCommand"] + "\n"

	dockerFile := getPhpImage +
		environmentInstallCmd +
		copyCommand +
		projectInstallCmd +
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
