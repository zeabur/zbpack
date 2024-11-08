# syntax=docker/dockerfile:1

ARG PHP_VERSION
FROM docker.io/library/php:${PHP_VERSION}-fpm

ENV APP_ENV=${APP_ENV:-prod}
ENV APP_DEBUG=${APP_DEBUG:-true}

WORKDIR /var/www

# install-php-extensions
ADD https://github.com/mlocati/docker-php-extension-installer/releases/latest/download/install-php-extensions /usr/local/bin/
RUN chmod +x /usr/local/bin/install-php-extensions && sync

# apt dependencies and node.js
ARG APT_EXTRA_DEPENDENCIES
RUN set -eux \
		&& apt update \
		&& apt install -y cron curl gettext git grep libicu-dev nginx pkg-config unzip ${APT_EXTRA_DEPENDENCIES} \
		&& rm -rf /var/www/html \
		&& curl -fsSL https://deb.nodesource.com/setup_22.x -o nodesource_setup.sh \
		&& bash nodesource_setup.sh \
		&& apt install -y nodejs \
		&& rm -rf /var/lib/apt/lists/*

# composer and php extensions
ARG PHP_EXTENSIONS
RUN install-php-extensions @composer apcu bcmath gd intl mysqli opcache pcntl pdo_mysql sysvsem zip ${PHP_EXTENSIONS}

# nginx configuration
RUN cat <<'EOF' > /etc/nginx/sites-enabled/default
server {
    listen 8080;
    root /var/www;

    add_header X-Frame-Options "SAMEORIGIN";
    add_header X-Content-Type-Options "nosniff";

    index index.php index.html;
    charset utf-8;

    location = /favicon.ico { access_log off; log_not_found off; }
    location = /robots.txt  { access_log off; log_not_found off; }

    error_page 404 /index.php;

    location ~ \.php$ {
        try_files $uri =404;
        fastcgi_split_path_info ^(.+\.php)(/.*)$;
        fastcgi_pass 127.0.0.1:9000;
        fastcgi_index index.php;
        include fastcgi_params;
        fastcgi_param SCRIPT_FILENAME $realpath_root$fastcgi_script_name;
        fastcgi_param DOCUMENT_ROOT $realpath_root;
        fastcgi_param PATH_INFO $fastcgi_path_info;
        fastcgi_hide_header X-Powered-By;
    }

    location / {
        try_files $uri $uri/ /index.php$is_args$args;
        gzip_static on;
    }

    location ~ /\.(?!well-known).* {
        deny all;
    }

    error_log /dev/stderr;
    access_log /dev/stderr;
}
EOF

# project directory
RUN chown -R www-data:www-data /var/www
COPY --link --chown=www-data:www-data --chmod=755 . /var/www

# install dependencies
USER www-data
RUN set -eux \
		&& if [ -f composer.json ]; then composer install --optimize-autoloader --classmap-authoritative --no-dev; fi \
		&& if [ -f package.json ]; then npm install; fi

ARG BUILD_COMMAND
RUN if [ -n "${BUILD_COMMAND}" ]; then ${BUILD_COMMAND}; fi

# optimization for frameworks
ARG PHP_OPTIMIZE
RUN <<EOF
	set -ux

	if [ ! "${PHP_OPTIMIZE}" = "true" ]; then
		echo "ZBPACK_PHP_OPTIMIZE is not set to true, skipping optimization"
		echo "You will need to run cache, optimization, and some build command manually."
		exit 0
	fi

	if [ -x artisan ]; then
		# Laravel
		php artisan optimize
		php artisan config:cache
		php artisan event:cache
		php artisan route:cache
		php artisan view:cache
	fi

	if [ -x bin/console ]; then
		# Symfony
		composer dump-env prod
		composer run-script --no-dev post-install-cmd
		php bin/console cache:clear

		# AssetMapper (optional)
		php bin/console asset-map:compile
	fi

	if [ -x ./node_modules/.bin/encore ]; then
		# Symfony Encore
		./node_modules/.bin/encore production
	fi

	# npm run build
	if grep -q '"build":' package.json; then
		npm run build
	fi
EOF

USER root

# if there is "public" directory in /var/www,
# we change the root directory to /var/www/public
RUN if [ -d /var/www/public ]; then sed -i 's|root /var/www;|root /var/www/public;|' /etc/nginx/sites-enabled/default; fi

ARG START_COMMAND
ENV START_COMMAND=${START_COMMAND}
CMD eval ${START_COMMAND}

EXPOSE 8080
