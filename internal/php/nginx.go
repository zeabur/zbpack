package php

import (
	"embed"
	"fmt"
	"io"
	"strings"

	"github.com/zeabur/zbpack/pkg/types"
)

//go:embed nginx-conf
var nginxConfFs embed.FS

// RetrieveNginxConf retrieves the nginx conf for the given app.
//
// The app should be an instance of types.PHPApplication; otherwise,
// an error will be returned.
func RetrieveNginxConf(app string) (string, error) {
	appConfMap := map[string]string{
		string(types.PHPApplicationDefault): "nginx-conf/default.conf",
		string(types.PHPApplicationAcgFaka): "nginx-conf/acg-faka.conf",
	}

	filename, ok := appConfMap[app]
	if !ok {
		return "", fmt.Errorf("unknown app: %s", app)
	}

	f, err := nginxConfFs.Open(filename)
	if err != nil {
		return "", fmt.Errorf("open: %w", err)
	}

	nginxConfBytes, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("readall: %w", err)
	}
	nginxConf := string(nginxConfBytes)

	return escape(nginxConf), nil
}

func escape(s string) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "$", "\\$")
	return s
}
