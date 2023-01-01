package php

import (
	"github.com/zeabur/zbpack/pkg/types"
)

func GenerateDockerfile(meta types.PlanMeta) (string, error) {
	phpVersion := meta["phpVersion"]
	getPhpImage := "FROM php:" + phpVersion + "\n"
	copyCmd := "COPY . .\n"
	//Get composer for next step install
	getComposerCmd := `RUN curl -sS https://getcomposer.org/installer | php && mv composer.phar /usr/local/bin/composer
`
	//The next step "composer install" will need to unzip some files
	getUnzipLibraryCmd := `RUN apt update && apt install unzip
`
	installCmd := "RUN composer install\n"
	startCmd := `CMD php artisan serve --port=8080`

	dockerFile := getPhpImage + copyCmd + getComposerCmd + getUnzipLibraryCmd + installCmd + startCmd
	return dockerFile, nil
}
