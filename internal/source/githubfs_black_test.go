package source_test

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeabur/zbpack/internal/source"
)

func getGithubToken(t *testing.T) *string {
	token, ok := os.LookupEnv("GITHUB_TOKEN")

	if !ok {
		t.Skip("no token (GITHUB_TOKEN) provided: skipping GitHub tests")
	}

	return &token
}

func TestGitHubFsOpen_File(t *testing.T) {
	token := getGithubToken(t)

	fs, err := source.NewGitHubFs("zeabur", "zeabur", token)
	require.NoError(t, err)

	t.Logf("fs: %#v", fs)

	f, err := fs.Open("readme.md")
	if err != nil {
		if strings.Contains(err.Error(), "401 Bad credentials") {
			t.Skip("Skip due to 401 error.")
			return
		}

		t.Fatalf("error when opening: %v", err)
	}

	content, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("error when reading file: %v", err)
	}

	t.Log(content)
}

func TestGitHubFsOpen_Dir(t *testing.T) {
	token := getGithubToken(t)

	fs, err := source.NewGitHubFs("zeabur", "zeabur", token)
	require.NoError(t, err)

	f, err := fs.Open("")
	if err != nil {
		if strings.Contains(err.Error(), "401 Bad credentials") {
			t.Skip("Skip due to 401 error.")
			return
		}

		t.Fatal("error when opening directory:", err)
	}

	fileInfo, err := f.Readdir(-1)
	if err != nil {
		t.Fatal("error when reading directory:", err)
	}

	for _, fi := range fileInfo {
		t.Log(fi.Name())
	}
}

func TestGitHubFsOpen_WithoutToken(t *testing.T) {
	// prevent rate limiting
	_ = getGithubToken(t)

	_, err := source.NewGitHubFs("zeabur", "zeabur", nil)
	assert.NoError(t, err)
}

func TestGitHubFsOpenFile_WithWriteFlag(t *testing.T) {
	token := getGithubToken(t)

	fs, err := source.NewGitHubFs("zeabur", "zeabur", token)
	require.NoError(t, err)

	_, err = fs.OpenFile("readme.md", os.O_RDWR, 0)
	assert.ErrorIs(t, err, os.ErrPermission)
}

func TestGitHubFsOpenFile_Ref(t *testing.T) {
	token := getGithubToken(t)

	fs, err := source.NewGitHubFs("zeabur", "zbpack", token, source.GitHubRef("9da82d05f3123cdb76b25d36c40cd12581e4eb82"))
	require.NoError(t, err)

	f, err := fs.OpenFile("go.mod", os.O_RDONLY, 0)
	if err != nil {
		if strings.Contains(err.Error(), "401 Bad credentials") {
			t.Skip("Skip due to 401 error.")
			return
		}

		t.Fatal("error when opening file:", err)
	}

	content, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("error when reading file: %v", err)
	}

	if !strings.Contains(string(content), "go 1.19") {
		t.Fatalf("unexpected content: %s", string(content))
	}

	t.Log(content)
}

func TestGitHubFsOpen_RefWithBranch(t *testing.T) {
	token := getGithubToken(t)

	fs, err := source.NewGitHubFs("zeabur", "zeabur", token, source.GitHubRef("marketplace-list"))
	require.NoError(t, err)

	f, err := fs.Open("marketplace.json")
	if err != nil {
		if strings.Contains(err.Error(), "401 Bad credentials") {
			t.Skip("Skip due to 401 error.")
			return
		}

		t.Fatalf("error when opening: %v", err)
	}

	content, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("error when reading file: %v", err)
	}

	assert.Contains(t, string(content), "codes")

	t.Log(content)
}

func TestGitHubFsOpen_Dir_Ref(t *testing.T) {
	token := getGithubToken(t)

	fs, err := source.NewGitHubFs("zeabur", "zeabur", token, source.GitHubRef("main"))
	require.NoError(t, err)

	f, err := fs.Open("")
	if err != nil {
		if strings.Contains(err.Error(), "401 Bad credentials") {
			t.Skip("Skip due to 401 error.")
			return
		}

		t.Fatal("error when opening directory:", err)
	}

	fileInfo, err := f.Readdir(-1)
	if err != nil {
		t.Fatal("error when reading directory:", err)
	}

	for _, fi := range fileInfo {
		t.Log(fi.Name())
	}
}

func TestGithubFs_Folder(t *testing.T) {
	token := getGithubToken(t)

	fs, err := source.NewGitHubFs("naiba", "nezha", token)
	require.NoError(t, err)

	f, err := fs.Open("")
	require.NoError(t, err)

	_, err = f.Stat()
	require.NoError(t, err)
}

func TestReadLimited(t *testing.T) {
	t.Parallel()

	t.Run("not oversized", func(t *testing.T) {
		r := strings.NewReader("hello, world")

		b, n, err := source.ReadLimited(r, 1024)
		require.NoError(t, err)
		assert.Equal(t, int64(12), n)
		assert.Equal(t, "hello, world", string(b))
	})

	t.Run("oversized", func(t *testing.T) {
		r := strings.NewReader(strings.Repeat("a", 1025))

		_, _, err := source.ReadLimited(r, 1024)
		assert.ErrorIs(t, err, source.ErrOverSized)
	})
}
