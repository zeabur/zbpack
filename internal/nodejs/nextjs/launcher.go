package nextjs

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"
)

//go:embed launcher.js.tmpl
var launcherTemplate string

// getNextConfig read .next/required-server-files.json and return the config string that will be injected into launcher
func getNextConfig() (string, error) {
	rsf, err := os.ReadFile(".next/required-server-files.json")
	if err != nil {
		return "", fmt.Errorf("read required-server-files.json: %w", err)
	}

	type requiredServerFiles struct {
		Config json.RawMessage `json:"config"`
	}

	var rs requiredServerFiles
	err = json.Unmarshal(rsf, &rs)
	if err != nil {
		return "", fmt.Errorf("unmarshal required-server-files.json: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(rs.Config, &data); err != nil {
		return "", fmt.Errorf("unmarshal config: %w", err)
	}

	nextConfig, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("marshal config: %w", err)
	}

	return string(nextConfig), nil
}

// renderLauncher will render the launcher.js template which used as the entrypoint of the serverless function
func renderLauncher() (string, error) {
	nextConfig, err := getNextConfig()
	if err != nil {
		return "", fmt.Errorf("get next config: %w", err)
	}

	tmpl, err := template.New("launcher").Parse(launcherTemplate)
	if err != nil {
		return "", fmt.Errorf("parse launcher template: %w", err)
	}

	type renderLauncherTemplateContext struct {
		NextConfig string
	}

	var launcher strings.Builder
	err = tmpl.Execute(&launcher, renderLauncherTemplateContext{NextConfig: nextConfig})
	if err != nil {
		return "", fmt.Errorf("render launcher template: %w", err)
	}

	return launcher.String(), nil
}
