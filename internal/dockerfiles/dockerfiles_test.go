package dockerfiles_test

import (
	"strings"
	"testing"

	"github.com/zeabur/zbpack/internal/dockerfiles"
)

func TestUnknownImage(t *testing.T) {
	t.Parallel()

	_, err := dockerfiles.GetDockerfileContent("unknown")
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetDockerfileContent(t *testing.T) {
	t.Parallel()

	content, err := dockerfiles.GetDockerfileContent("test")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), `LABEL com.zeabur.image-type="test"`) {
		t.Errorf("missing label, content=%s", string(content))
	}
}
