package php

import (
	"github.com/zeabur/zbpack/pkg/types"
)

func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	phpVersion := meta["phpVersion"]
	getPhpImage := "FROM php:" + phpVersion + "-fpm\n"
	// environment command: apt-lib + php-extenstion + composer
	envInstallCmd := `
	RUN apt-get update && apt-get install -y \
    build-essential \
    libpng-dev \
    libjpeg62-turbo-dev \
    libfreetype6-dev \
    locales \
    zip \
    jpegoptim optipng pngquant gifsicle \
    unzip \
    git \
    curl \
    lua-zlib-dev \
    libmemcached-dev \
    nginx
	RUN apt-get install -y supervisor
	RUN curl -sS https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer
	RUN apt-get clean && rm -rf /var/lib/apt/lists/*`
	// copy source code
	copyCommand := `
	COPY --chown=www-data:www-data . /var/www  
	COPY ./docker/supervisor.conf /etc/supervisord.conf
	COPY ./docker/php.ini /usr/local/etc/php/conf.d/app.ini
	COPY ./docker/nginx.conf /etc/nginx/sites-enabled/default`

	switch meta["framework"] {
	case "laravel":
		copyCommand = `
		COPY --chown=www-data:www-data . /var/www  
		RUN chmod -R 755 /var/www/storage
		COPY ./docker/supervisor.conf /etc/supervisord.conf
		COPY ./docker/php.ini /usr/local/etc/php/conf.d/app.ini
		COPY ./docker/nginx.conf /etc/nginx/sites-enabled/default`
	}
	// PHP Error Log Files
	phpErrLog := `
	RUN mkdir /var/log/php
	RUN touch /var/log/php/errors.log && chmod 777 /var/log/php/errors.log`
	//composer install
	composerInstallCmd := `
	RUN composer install --optimize-autoloader --no-dev`

	//startCmd
	startCmd := `
	EXPOSE ${PORT}
	CMD ["php","artisan","cache:clear"]
	CMD ["php","artisan","route:cache"]
	CMD ["/usr/bin/supervisord","-c","/etc/supervisord.conf"]
	`
	dockerFile := getPhpImage + `WORKDIR /var/www
	ADD https://github.com/mlocati/docker-php-extension-installer/releases/latest/download/install-php-extensions /usr/local/bin/
	RUN chmod +x /usr/local/bin/install-php-extensions && sync && \
    install-php-extensions mbstring pdo_mysql zip exif pcntl gd memcached` +
		envInstallCmd +
		copyCommand +
		phpErrLog +
		composerInstallCmd +
		startCmd
	return dockerFile, nil
}
