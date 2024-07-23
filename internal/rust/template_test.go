package rust_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/internal/rust"
)

func TestGenerateDockerfile_Assets(t *testing.T) {
	t.Parallel()

	t.Run("one assets", func(t *testing.T) {
		t.Parallel()

		meta := map[string]string{
			"openssl":    "true",
			"serverless": "true",
			"entry":      "entry",
			"appDir":     "appDir",
			"assets":     "assets",
		}

		dockerfile, err := rust.GenerateDockerfile(meta)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Contains(t, dockerfile, "COPY --from=builder /app/assets /app/assets")
	})

	t.Run("multiple assets", func(t *testing.T) {
		t.Parallel()

		meta := map[string]string{
			"openssl":    "true",
			"serverless": "true",
			"entry":      "entry",
			"appDir":     "appDir",
			"assets":     "assets1:assets2",
		}

		dockerfile, err := rust.GenerateDockerfile(meta)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Contains(t, dockerfile, "COPY --from=builder /app/assets1 /app/assets1")
		assert.Contains(t, dockerfile, "COPY --from=builder /app/assets2 /app/assets2")
	})
}

func TestGenerateDockerfile_OpenSSL(t *testing.T) {
	t.Parallel()

	t.Run("true", func(t *testing.T) {
		t.Parallel()

		meta := map[string]string{
			"openssl":    "true",
			"serverless": "false",
			"entry":      "entry",
			"appDir":     "appDir",
			"assets":     "assets",
		}

		dockerfile, err := rust.GenerateDockerfile(meta)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Contains(t, dockerfile, "apt-get install -y openssl")
	})
}

func TestGenerateDockerfile_Serverless(t *testing.T) {
	t.Parallel()

	t.Run("true", func(t *testing.T) {
		t.Parallel()

		meta := map[string]string{
			"openssl":    "false",
			"serverless": "true",
			"entry":      "entry",
			"appDir":     "appDir",
			"assets":     "assets",
		}

		dockerfile, err := rust.GenerateDockerfile(meta)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Contains(t, dockerfile, "FROM scratch")
		assert.Contains(t, dockerfile, "COPY --from=post-builder /app .")
	})

	t.Run("false", func(t *testing.T) {
		t.Parallel()

		meta := map[string]string{
			"openssl":    "false",
			"serverless": "false",
			"entry":      "entry",
			"appDir":     "appDir",
			"assets":     "assets",
		}

		dockerfile, err := rust.GenerateDockerfile(meta)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Contains(t, dockerfile, "FROM rust:1-slim AS runtime")
		assert.Contains(t, dockerfile, "COPY --from=post-builder /app /app")
		assert.Contains(t, dockerfile, `CMD ["/app/main"]`)
	})
}

func TestGenerateDockerfile_AppDir(t *testing.T) {
	t.Parallel()

	t.Run("configured", func(t *testing.T) {
		t.Parallel()

		meta := map[string]string{
			"openssl":    "false",
			"serverless": "false",
			"entry":      "entry",
			"appDir":     "configured",
			"assets":     "assets",
		}

		dockerfile, err := rust.GenerateDockerfile(meta)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Contains(t, dockerfile, `cargo install --path "configured" --root /out`)
	})

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		meta := map[string]string{
			"openssl":    "false",
			"serverless": "false",
			"entry":      "entry",
			"appDir":     ".",
			"assets":     "assets",
		}

		dockerfile, err := rust.GenerateDockerfile(meta)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Contains(t, dockerfile, `cargo install --path "." --root /out`)
	})
}

func TestGenerateDockerfile_Entry(t *testing.T) {
	t.Parallel()

	t.Run("configured", func(t *testing.T) {
		t.Parallel()

		meta := map[string]string{
			"openssl":    "false",
			"serverless": "false",
			"entry":      "configured",
			"appDir":     ".",
			"assets":     "assets",
		}

		dockerfile, err := rust.GenerateDockerfile(meta)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Contains(t, dockerfile, `if [ -x "configured" ]; then`)
		assert.Contains(t, dockerfile, `mv "configured" /app/main`)
	})
}

func TestGenerateDockerfile_Workdir(t *testing.T) {
	t.Parallel()

	meta := map[string]string{
		"openssl":    "false",
		"serverless": "false",
		"entry":      "entry",
		"appDir":     ".",
		"assets":     "assets",
	}

	dockerfile, err := rust.GenerateDockerfile(meta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assert.Contains(t, dockerfile, `WORKDIR /app`)
}
