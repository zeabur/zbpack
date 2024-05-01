package php

import (
	"strings"

	"github.com/spf13/afero"
)

func determineSailRuntime(source afero.Fs) string {
	runtime := "./vendor/laravel/sail/runtimes/8.3"

	compose, err := afero.ReadFile(source, "docker-compose.yml")
	if err == nil && strings.Contains(string(compose), "vendor/laravel/sail/runtimes") {
		lines := strings.Split(string(compose), "\n")
		for _, line := range lines {
			if strings.Contains(line, "vendor/laravel/sail/runtimes") {
				parts := strings.Split(line, "/")
				runtime = "./vendor/laravel/sail/runtimes/" + parts[len(parts)-1]
			}
		}
	}

	dockerfile, err := afero.ReadFile(source, runtime+"/Dockerfile")
	if err != nil {
		return ""
	}

	lines := strings.Split(string(dockerfile), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "COPY") {
			parts := strings.Split(line, " ")
			parts[1] = runtime + "/" + parts[1]
			lines[i] = strings.Join(parts, " ")
		}

		if strings.Contains(line, "$WWWGROUP") {
			lines[i] = strings.ReplaceAll(line, "$WWWGROUP", "1000")
		}

		if strings.HasPrefix(line, "ENTRYPOINT") {
			lines[i] = `COPY . /var/www/html
RUN chown -R sail:sail /var/www/html
` + line
		}

		if strings.Contains(line, "EXPOSE") {
			lines[i] = ""
		}
	}

	return strings.Join(lines, "\n")
}
