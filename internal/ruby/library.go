package ruby

import (
	"fmt"
	"github.com/zeabur/zbpack/internal/source"
	"regexp"
	"strings"
)

func GetGemfileValue(source *source.Source, key string) string {
	src := *source
	var ret string
	file, err := src.ReadFile("Gemfile")
	if err != nil {
		fmt.Errorf("failed to parse Gemfile: %w", err)
		return ""
	}
	matchString := regexp.MustCompile(key)
	lines := strings.Split(string(file), "\n")
	for _, line := range lines {
		if matchString.Match([]byte(line)) {
			ret = strings.TrimPrefix(line, key)
			return ret
		}
	}
	return ""
}
