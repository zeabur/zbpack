package utils

import (
	"testing"
)

func TestExtractErlangEntryFromGleamEntrypointShell(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input: `#!/bin/sh
set -eu

PACKAGE=my_web_app
BASE=$(dirname "$0")
COMMAND="${1-default}"

run() {
  erl \
    -pa "$BASE"/*/ebin \
    -eval "$PACKAGE@@main:run($PACKAGE)" \
    -noshell \
    -extra "$@"
}

shell() {
  erl -pa "$BASE"/*/ebin
}

case "$COMMAND" in
  run)
    shift
    run "$@"
  ;;

  shell)
    shell
  ;;

  *)
    echo "usage:" >&2
    echo "  entrypoint.sh \$COMMAND" >&2
    echo "" >&2
    echo "commands:" >&2
    echo "  run    Run the project main function" >&2
    echo "  shell  Run an Erlang shell" >&2
    exit 1
esac`,
			expected: "my_web_app@@main:run(my_web_app)",
		},
		{
			input: `#!/bin/sh
set -eu

PACKAGE=another_app
BASE=$(dirname "$0")
COMMAND="${1-default}"

run() {
  erl \
    -pa "$BASE"/*/ebin \
    -eval "$PACKAGE@@another:action($PACKAGE)" \
    -noshell \
    -extra "$@"
}

shell() {
  erl -pa "$BASE"/*/ebin
}

case "$COMMAND" in
  run)
    shift
    run "$@"
  ;;

  shell)
    shell
  ;;

  *)
    echo "usage:" >&2
    echo "  entrypoint.sh \$COMMAND" >&2
    echo "" >&2
    echo "commands:" >&2
    echo "  run    Run the project main function" >&2
    echo "  shell  Run an Erlang shell" >&2
    exit 1
esac`,
			expected: "another_app@@another:action(another_app)",
		},
		{
			input: `#!/bin/sh
set -eu

PACKAGE=no_eval
BASE=$(dirname "$0")
COMMAND="${1-default}"

run() {
  erl \
    -pa "$BASE"/*/ebin \
    -noshell \
    -extra "$@"
}

shell() {
  erl -pa "$BASE"/*/ebin
}

case "$COMMAND" in
  run)
    shift
    run "$@"
  ;;

  shell)
    shell
  ;;

  *)
    echo "usage:" >&2
    echo "  entrypoint.sh \$COMMAND" >&2
    echo "" >&2
    echo "commands:" >&2
    echo "  run    Run the project main function" >&2
    echo "  shell  Run an Erlang shell" >&2
    exit 1
esac`,
			expected: "",
		},
	}

	for _, test := range tests {
		result := ExtractErlangEntryFromGleamEntrypointShell(test.input)
		if result != test.expected {
			t.Errorf("For input:\n%s\nexpected: %s\nbut got: %s", test.input, test.expected, result)
		}
	}
}
