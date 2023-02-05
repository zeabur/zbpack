package ruby

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
)

func GetGemfileValue(absPath string, key string) string {
	filePath := path.Join(absPath, "Gemfile")
	var ret string
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Errorf("failed to parse Gemfile: %w", err)
	}
	defer file.Close()
	matchString := regexp.MustCompile(key)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := []byte(scanner.Text())
		if matchString.Match(line) {
			ret = strings.Trim(scanner.Text(), key)
			return ret
		}
	}
	return ""
}
