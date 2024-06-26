package php

import (
	"encoding/json"

	"github.com/spf13/afero"
	"github.com/zeabur/zbpack/internal/utils"
)

type composerJSONSchema struct {
	Name       string             `json:"name"`
	Require    *map[string]string `json:"require"`
	RequireDev *map[string]string `json:"require-dev"`
}

func (c *composerJSONSchema) GetRequire(dep string) (string, bool) {
	if c.Require == nil {
		return "", false
	}
	v, ok := (*c.Require)[dep]
	return v, ok
}

func (c *composerJSONSchema) GetRequireDev(dep string) (string, bool) {
	if c.RequireDev == nil {
		return "", false
	}
	v, ok := (*c.RequireDev)[dep]
	return v, ok
}

func parseComposerJSON(source afero.Fs) (composerJSONSchema, error) {
	composerJSONMarshal, err := utils.ReadFileToUTF8(source, "composer.json")
	if err != nil {
		return composerJSONSchema{}, err
	}
	var composerJSON composerJSONSchema

	if err := json.Unmarshal(composerJSONMarshal, &composerJSON); err != nil {
		return composerJSONSchema{}, err
	}

	return composerJSON, nil
}
