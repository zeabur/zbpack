package zeaburpack

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/pan93412/envexpander/v3"
	"github.com/samber/mo"
)

// InjectDockerfile injects the environment variables and
// the Docker.io registry into the Dockerfile.
func InjectDockerfile(dockerfile string, registry string, variables map[string]string) string {
	// resolve env variable statically and don't depend on Dockerfile's order
	resolvedVars := envexpander.Expand(variables)

	refConstructor := newReferenceConstructor(&registry)
	lines := strings.Split(dockerfile, "\n")
	stageLines := make([]int, 0)

	for i, line := range lines {
		fromStatement, isFromStatement := ParseFrom(line)
		if !isFromStatement {
			continue
		}

		// Construct the reference.
		newRef := refConstructor.Construct(fromStatement.Source)

		// Replace this FROM line.
		fromStatement.Source = newRef
		lines[i] = fromStatement.String()

		// Mark this FROM line as a stage.
		if stage, ok := fromStatement.Stage.Get(); ok {
			refConstructor.AddStage(stage)
		}
		stageLines = append(stageLines, i)
	}

	// sort the resolvedVars by key so we can build
	// the reproducible dockerfile
	sortedResolvedVarsKey := make([]string, 0, len(resolvedVars))
	for key := range resolvedVars {
		sortedResolvedVarsKey = append(sortedResolvedVarsKey, key)
	}
	sort.Strings(sortedResolvedVarsKey)

	// build the dockerfile
	dockerfileEnv := ""

	for _, key := range sortedResolvedVarsKey {
		value := strconv.Quote(resolvedVars[key])
		dockerfileEnv += fmt.Sprintf(`ENV %s="%s"`, key, strconv.Quote(value)) + "\n"
	}

	for _, stageLine := range stageLines {
		lines[stageLine] = lines[stageLine] + "\n" + dockerfileEnv + "\n"
	}

	return strings.Join(lines, "\n")
}

// FromStatement represents a FROM statement in a Dockerfile.
type FromStatement struct {
	Source string
	Stage  mo.Option[string]
}

// ParseFrom parses a FROM statement from a Dockerfile line.
func ParseFrom(line string) (FromStatement, bool) {
	parsed, err := parser.Parse(strings.NewReader(line))
	if err != nil {
		return FromStatement{}, false
	}

	for _, child := range parsed.AST.Children {
		if child.Value == "FROM" {
			source := child.Next.Value

			if child.Next.Next != nil && child.Next.Next.Value == "AS" {
				return FromStatement{
					Source: child.Next.Value,
					Stage:  mo.Some(child.Next.Next.Next.Value),
				}, true
			}

			return FromStatement{
				Source: source,
				Stage:  mo.None[string](),
			}, true
		}
	}

	return FromStatement{}, false
}

func (fs FromStatement) String() string {
	if stage, ok := fs.Stage.Get(); ok {
		return "FROM " + fs.Source + " AS " + stage
	}

	return "FROM " + fs.Source
}

// GetImageType extracts the image type of the specified stage from the Dockerfile.
func GetImageType(dockerfile string, stage string) string {
	reader := bytes.NewReader([]byte(dockerfile))
	parsed, err := parser.Parse(reader)
	if err != nil {
		return ""
	}

	var finalStage string
	var currentStage string
	imageType := ""

	for _, child := range parsed.AST.Children {
		if child.Value == "FROM" {
			// Handle stages
			currentStage = ""
			if child.Next != nil && child.Next.Next != nil && child.Next.Next.Value == "AS" {
				currentStage = child.Next.Next.Next.Value
			}
			finalStage = currentStage // Always track the final stage
		}

		if child.Value == "LABEL" && child.Next != nil && child.Next.Value == "com.zeabur.image-type" {
			if currentStage == stage || (stage == "" && currentStage == finalStage) {
				imageType, _ = strconv.Unquote(child.Next.Next.Value)
			}
		}
	}

	return imageType
}
