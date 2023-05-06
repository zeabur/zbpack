package ruby

import (
	"log"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

func GetGemfileValue(source afero.Fs, key string) string {
	var ret string
	file, err := afero.ReadFile(source, "Gemfile")
	if err != nil {
		// TODO)) return
		log.Printf("failed to parse Gemfile: %v", err)
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
