package zeaburpack

import (
	"regexp"

	"github.com/samber/mo"
)

var fromStatementRegex = regexp.MustCompile(`(?i)^\s*FROM\s+(?P<src>\S*)(?:\s+AS\s+(?P<stage>\S+))?`)

// FromStatement represents a FROM statement in a Dockerfile.
type FromStatement struct {
	Source string
	Stage  mo.Option[string]
}

// ParseFrom parses a FROM statement from a Dockerfile line.
func ParseFrom(line string) (FromStatement, bool) {
	statement := FromStatement{}

	names := fromStatementRegex.SubexpNames()
	submatch := fromStatementRegex.FindStringSubmatch(line)
	if len(submatch) == 0 {
		return statement, false
	}

	for i, name := range names {
		switch name {
		case "src":
			statement.Source = submatch[i]
		case "stage":
			if submatch[i] != "" {
				statement.Stage = mo.Some(submatch[i])
			}
		}
	}

	return statement, true
}

func (fs FromStatement) String() string {
	if stage, ok := fs.Stage.Get(); ok {
		return "FROM " + fs.Source + " AS " + stage
	}

	return "FROM " + fs.Source
}
