package php

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	. "github.com/zeabur/zbpack/pkg/types"
)

func GetPhpVersion(absPath string) string {
	composerJsonMarshal, err := os.ReadFile(path.Join(absPath, "composer.json"))
	if err != nil {
		return ""
	}
	composerJson := struct {
		Require map[string]string `json:"require"`
	}{}

	if err := json.Unmarshal(composerJsonMarshal, &composerJson); err != nil {
		return "8.0"
	}
	if composerJson.Require["php"] == "" {
		return "8.0"
	}

	// for example, ">=16.0.0 <17.0.0"
	// versionRange := composerJson.Engines.Node

	// isVersion, _ := regexp.MatchString(`^\d+(\.\d+){0,2}$`, versionRange)
	// if isVersion {
	// 	return versionRange
	// }
	versionRange := composerJson.Require["php"]

	isVersion, _ := regexp.MatchString(`^\d+(\.\d+){0,2}$`, versionRange)
	if isVersion {
		return versionRange
	}
	ranges := strings.Split(versionRange, " ")
	// equalMin := false
	// maxVer := -1
	// equalMax := false
	for _, r := range ranges {
		if strings.HasPrefix(r, ">=") {
			minVerString := strings.TrimPrefix(r, ">=")
			return minVerString
		} else if strings.HasPrefix(r, ">") {
			minVerString := strings.TrimPrefix(r, ">")
			value, err := strconv.ParseFloat(minVerString, 64)
			if err != nil {
				// insert error handling here
			}
			value += 0.1
			minVerString = fmt.Sprintf("%f", value)
			// equalMin = false
			return minVerString
		} else if strings.HasPrefix(r, "<=") {
			maxVerString := strings.TrimPrefix(r, "<=")
			return maxVerString

		} else if strings.HasPrefix(r, "<") {
			maxVerString := strings.TrimPrefix(r, "<=")
			value, err := strconv.ParseFloat(maxVerString, 64)
			if err != nil {
				// insert error handling here
			}
			value -= 0.1

			maxVerString = fmt.Sprintf("%f", value)
			return maxVerString
		}
	}

	return "8.1"
}

func WriteConfigForPhpImage(absPath string) bool {
	path := absPath + "docker/"
	err := os.Mkdir(path, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}
	nginxFile, err := os.Create(path + "nginx.conf")
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer nginxFile.Close()
	nginxFile.WriteString(`server {
	listen 8080;
	root /var/www/public;

	add_header X-Frame-Options "SAMEORIGIN";
	add_header X-Content-Type-Options "nosniff";

	index index.php index.html;
	charset utf-8;

	location = /favicon.ico { access_log off; log_not_found off; }
	location = /robots.txt  { access_log off; log_not_found off; }

	error_page 404 /index.php;

	location ~ \.php$ {
		try_files $uri =404;
		fastcgi_split_path_info ^(.+\.php)(/.+)$;
		fastcgi_pass 127.0.0.1:9000;
		fastcgi_index index.php;
		include fastcgi_params;
		fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
		fastcgi_param PATH_INFO $fastcgi_path_info;
		fastcgi_buffering off;
	}

	location / {
		try_files $uri $uri/ /index.php?$query_string;
		gzip_static on;
	}
	
	location ~ /\.(?!well-known).* {
		deny all;
	}
	}`)
	supervisorFile, err := os.Create(path + "supervisor.conf")
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer supervisorFile.Close()
	supervisorFile.WriteString(
		`[supervisord]
nodaemon=true
loglevel = info
logfile=/var/log/supervisord.log
pidfile=/var/run/supervisord.pid

[group:laravel-worker]
priority=999
programs=nginx,php8-fpm,laravel-schedule,laravel-notification,laravel-queue

[program:nginx]
priority=10
autostart=true
autorestart=true
stderr_logfile_maxbytes=0
stdout_logfile_maxbytes=0
stdout_events_enabled=true
stderr_events_enabled=true
command=/usr/sbin/nginx -g 'daemon off;'
stderr_logfile=/var/log/nginx/error.log
stdout_logfile=/var/log/nginx/access.log

[program:php8-fpm]
priority=5
autostart=true
autorestart=true
stderr_logfile_maxbytes=0
stdout_logfile_maxbytes=0
command=/usr/local/sbin/php-fpm -R
stderr_logfile=/var/log/nginx/php-error.log
stdout_logfile=/var/log/nginx/php-access.log

[program:laravel-schedule]
numprocs=1
autostart=true
autorestart=true
redirect_stderr=true
process_name=%(program_name)s_%(process_num)02d
command=php /var/www/artisan schedule:run
stdout_logfile=/var/log/nginx/schedule.log

[program:laravel-notification]
numprocs=1
autostart=true
autorestart=true
redirect_stderr=true
process_name=%(program_name)s_%(process_num)02d
command=php /var/www/artisan notification:worker
stdout_logfile=/var/log/nginx/notification.log

[program:laravel-queue]
numprocs=5
autostart=true
autorestart=true
redirect_stderr=true
process_name=%(program_name)s_%(process_num)02d
stdout_logfile=/var/log/nginx/worker.log
command=php /var/www/artisan queue:work sqs --sleep=3 --tries=3`)

	phpiniFile, err := os.Create(path + "php.ini")
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer phpiniFile.Close()
	phpiniFile.WriteString(`
log_errors=1
display_errors=1
post_max_size=40M
upload_max_filesize=40M
display_startup_errors=1
error_log=/var/log/php/errors.log`)
	return true
}

func DetermineProjectFramework(absPath string) PhpFramework {
	composerJsonMarshal, err := os.ReadFile(path.Join(absPath, "composer.json"))

	WriteConfigForPhpImage(absPath)

	if err != nil {
		return PhpFrameworkNone
	}
	composerJson := struct {
		Name       string            `json:"name"`
		Require    map[string]string `json:"require"`
		Requiredev map[string]string `json:"require-dev"`
	}{}
	if err := json.Unmarshal(composerJsonMarshal, &composerJson); err != nil {
		return PhpFrameworkNone
	}
	if _, isLaravel := composerJson.Require["laravel/framework"]; isLaravel {
		return PhpFrameworkLaravel
	}

	return PhpFrameworkNone

}
