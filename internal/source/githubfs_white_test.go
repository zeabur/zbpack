package source

import (
	"bytes"
	"io"
	"testing"

	"github.com/google/go-github/v45/github"
	"github.com/stretchr/testify/assert"
)

func TestGitHubFile_ReadFile(t *testing.T) {
	file := &githubFile{
		Reader: bytes.NewReader([]byte("hello world")),
		info: &githubFileInfo{
			name:  "readme.md",
			isDir: false,
		},
	}

	buf := make([]byte, 1)
	n, err := file.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, []byte("h"), buf)
}

func TestGitHubFile_ReadDir_ThrowNotDir(t *testing.T) {
	var f githubFile

	_, err := f.Readdir(1)
	assert.ErrorIs(t, err, ErrNotDir)
}

func TestGitHubFile_Create_ThrowReadonly(t *testing.T) {
	var f githubFile

	_, err := f.WriteString("hi")
	assert.ErrorIs(t, err, ErrReadonly)
}

func TestGitHubDir_ReadDir_NegativeReturnsAll(t *testing.T) {
	dir := &githubDir{
		githubFile: &githubFile{
			Reader: &bytes.Reader{},
			info: &githubFileInfo{
				name:  "dir",
				isDir: true,
			},
		},
		contents: []*github.RepositoryContent{
			{
				Name: github.String("readme.md"),
			},
		},
	}

	info, err := dir.Readdir(-1)
	assert.NoError(t, err)
	assert.Len(t, info, 1)
	assert.Equal(t, "readme.md", info[0].Name())
}

func TestGitHubDir_ReadDir_ZeroReturnsAll(t *testing.T) {
	dir := &githubDir{
		githubFile: &githubFile{
			Reader: &bytes.Reader{},
			info: &githubFileInfo{
				name:  "dir",
				isDir: true,
			},
		},
		contents: []*github.RepositoryContent{
			{
				Name: github.String("readme.md"),
			},
		},
	}

	info, err := dir.Readdir(0)
	assert.NoError(t, err)
	assert.Len(t, info, 1)
}

func TestGitHubDir_ReadDir_PositiveReturnMax(t *testing.T) {
	// when n > max, return max

	dir := &githubDir{
		githubFile: &githubFile{
			Reader: &bytes.Reader{},
			info: &githubFileInfo{
				name:  "dir",
				isDir: true,
			},
		},
		contents: []*github.RepositoryContent{
			{
				Name: github.String("readme.md"),
			},
		},
	}

	info, err := dir.Readdir(114514)
	assert.NoError(t, err)
	assert.Len(t, info, 1)

	_, err = dir.Readdir(222222)
	assert.Error(t, err, io.EOF)
}

func TestGitHubDir_ReadDir_PositiveReturnN(t *testing.T) {
	// when n > max, return max

	dir := &githubDir{
		githubFile: &githubFile{
			Reader: &bytes.Reader{},
			info: &githubFileInfo{
				name:  "dir",
				isDir: true,
			},
		},
		contents: []*github.RepositoryContent{
			{
				Name: github.String("readme.md"),
			},
			{
				Name: github.String("readme2.md"),
			},
		},
	}

	info, err := dir.Readdir(1)
	assert.NoError(t, err)
	assert.Len(t, info, 1)
	assert.True(t, info[0].Name() == "readme.md")

	info, err = dir.Readdir(1)
	assert.NoError(t, err)
	assert.Len(t, info, 1)
	assert.Equal(t, "readme2.md", info[0].Name())

	_, err = dir.Readdir(1)
	assert.Error(t, err, io.EOF)
}

func TestGitHubDir_ReadDir_Negative_WithoutContent(t *testing.T) {
	dir := &githubDir{
		githubFile: &githubFile{
			Reader: &bytes.Reader{},
			info: &githubFileInfo{
				name:  "dir",
				isDir: true,
			},
		},
		contents: nil,
	}

	info, err := dir.Readdir(-1)
	assert.NoError(t, err)
	assert.Len(t, info, 0)
}

func TestGitHubDir_ReadDir_Positive_WithoutContent(t *testing.T) {
	dir := &githubDir{
		githubFile: &githubFile{
			Reader: &bytes.Reader{},
			info: &githubFileInfo{
				name:  "dir",
				isDir: true,
			},
		},
		contents: nil,
	}

	_, err := dir.Readdir(1)
	assert.Error(t, err, io.EOF)
}

func TestGitHubDir_ReadDirnames(t *testing.T) {
	dir := &githubDir{
		githubFile: &githubFile{
			Reader: &bytes.Reader{},
			info: &githubFileInfo{
				name:  "dir",
				isDir: true,
			},
		},
		contents: []*github.RepositoryContent{
			{
				Name: github.String("readme.md"),
			},
			{
				Name: github.String("readme2.md"),
			},
		},
	}

	info, err := dir.Readdirnames(-1)
	assert.NoError(t, err)
	assert.Len(t, info, 2)

	assert.Equal(t, "readme.md", info[0])
	assert.Equal(t, "readme2.md", info[1])

	_, err = dir.Readdirnames(1)
	assert.Error(t, err, io.EOF)
}
