package zeaburpack

import (
	"strconv"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/samber/mo"
)

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
		if strings.ToUpper(child.Value) == "FROM" {
			source := child.Next.Value

			if child.Next.Next != nil && strings.ToUpper(child.Next.Next.Value) == "AS" {
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

// ExtractLabels extracts the labels from Dockerfile.
//
// Note that it picks the final label if there are multiple labels with the same key.
func ExtractLabels(dockerfile string) map[string]string {
	labels := make(map[string]string)

	parsed, err := parser.Parse(strings.NewReader(dockerfile))
	if err != nil {
		return labels
	}

	for _, child := range parsed.AST.Children {
		if strings.ToUpper(child.Value) == "LABEL" {
			currentLabelNode := child

			for currentLabelNode.Next != nil && currentLabelNode.Next.Next != nil {
				key, err := strconv.Unquote(currentLabelNode.Next.Value)
				if err != nil {
					key = currentLabelNode.Next.Value
				}
				value, err := strconv.Unquote(currentLabelNode.Next.Next.Value)
				if err != nil {
					value = currentLabelNode.Next.Next.Value
				}

				labels[key] = value

				if currentLabelNode.Next.Next.Next != nil {
					currentLabelNode = currentLabelNode.Next.Next.Next
				}
			}
		}
	}

	return labels
}

func (fs FromStatement) String() string {
	if stage, ok := fs.Stage.Get(); ok {
		return "FROM " + fs.Source + " AS " + stage
	}

	return "FROM " + fs.Source
}
