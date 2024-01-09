package types

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

// ZeaburOutputConfigRoute is a route in the output config to override the default route
// src is the path regex want to override, dest is the path you want to override it with
// for example, assume we already have an index.html in .zeabur/output/static,
// and our service is a Single Page App, we want to override all routes to serve index.html
// we would add the following to the output config:
// { "src": ".*", "dest": "/index.html" }
type ZeaburOutputConfigRoute struct {
	Src  string `json:"src"`
	Dest string `json:"dest"`
}

// ZeaburOutputConfig is the output config of Zeabur
type ZeaburOutputConfig struct {
	// Routes is a list of routes to override the default route
	Routes []ZeaburOutputConfigRoute `json:"routes"`
}

type ZeaburOutputFunctionConfig struct {
	Runtime string `json:"runtime"`
	Entry   string `json:"entry,omitempty"`
}

func (c ZeaburOutputFunctionConfig) WriteTo(funcPath string) error {
	funcConfigBytes, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal function config: %w", err)
	}

	err = os.WriteFile(path.Join(funcPath, ".zb-config.json"), funcConfigBytes, 0644)
	if err != nil {
		return fmt.Errorf("write function config: %w", err)
	}

	return nil
}
