package source

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/spf13/afero"
	"golang.org/x/oauth2"
)

var _ afero.Fs = githubFs{}

type githubFs struct {
	GitHubRepoOwner string
	GitHubRepoName  string
	GitHubClient    *github.Client
}

// NewGitHubFs creates a new github filesystem.
func NewGitHubFs(repoOwner, repoName, token string) afero.Fs {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	oauth2Client := oauth2.NewClient(context.TODO(), tokenSource)
	client := github.NewClient(oauth2Client)
	return githubFs{
		GitHubRepoName:  repoName,
		GitHubClient:    client,
		GitHubRepoOwner: repoOwner,
	}
}

var (
	// ErrUnimplemented is returned when a method is not implemented.
	ErrUnimplemented = errors.New("unimplemented")
	// ErrReadonly is returned when this filesystem is readonly.
	ErrReadonly = errors.New("readonly")
	// ErrNotDir is returned when something is not a directory.
	ErrNotDir = errors.New("not a directory")
)

func (fs githubFs) Create(string) (afero.File, error) {
	return nil, ErrReadonly
}

func (fs githubFs) Mkdir(string, os.FileMode) error {
	return ErrReadonly
}

func (fs githubFs) MkdirAll(string, os.FileMode) error {
	return ErrReadonly
}

func (fs githubFs) Open(name string) (afero.File, error) {
	root, dirContent, _, err := fs.GitHubClient.Repositories.GetContents(
		context.TODO(), fs.GitHubRepoOwner, fs.GitHubRepoName, name,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	if len(dirContent) > 0 {
		return fs.openAsDir(root, dirContent)
	}

	return fs.openAsFile(root)
}

func (fs githubFs) openAsFile(f *github.RepositoryContent) (afero.File, error) {
	c, err := f.GetContent()
	if err != nil {
		return nil, fmt.Errorf("read content: %w", err)
	}

	return &githubFile{
		Reader: bytes.NewReader([]byte(c)),
		info:   repoContentToFileInfo(f),
	}, nil
}

func (fs githubFs) openAsDir(root *github.RepositoryContent, content []*github.RepositoryContent) (afero.File, error) {
	return &githubDir{
		contents: content,
		githubFile: &githubFile{
			Reader: bytes.NewReader([]byte{}), // empty
			info:   repoContentToFileInfo(root),
		},
	}, nil
}

func (fs githubFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if flag != os.O_RDONLY {
		return nil, ErrReadonly
	}

	if perm != 0o444 {
		log.Println("GitHubFs only supports read-only files.")
		log.Println("Therefore, the perm argument is always 0444.")
	}

	return fs.Open(name)
}

func (fs githubFs) Remove(string) error {
	return ErrReadonly
}

func (fs githubFs) RemoveAll(string) error {
	return ErrReadonly
}

func (fs githubFs) Rename(_, _ string) error {
	return ErrReadonly
}

func (fs githubFs) Stat(name string) (os.FileInfo, error) {
	f, err := fs.Open(name)
	if err != nil {
		return nil, err
	}
	return f.Stat()
}

func (fs githubFs) Name() string {
	return "GithubFs"
}

func (fs githubFs) Chmod(_ string, _ os.FileMode) error {
	return ErrReadonly
}

func (fs githubFs) Chtimes(_ string, _, _ time.Time) error {
	return ErrReadonly
}

func (fs githubFs) Chown(_ string, _, _ int) error {
	return ErrReadonly
}

var _ afero.File = &githubFile{}

type githubFile struct {
	*bytes.Reader

	info os.FileInfo
}

func (f githubFile) Close() error {
	return nil // we don't need to close it
}

func (f githubFile) Name() string {
	return f.info.Name()
}

func (f githubFile) Readdir(int) ([]os.FileInfo, error) {
	return nil, ErrNotDir
}

func (f githubFile) Readdirnames(int) ([]string, error) {
	return nil, ErrNotDir
}

func (f githubFile) Stat() (os.FileInfo, error) {
	return f.info, nil
}

func (f githubFile) Sync() error {
	return nil // we don't need to sync a []byte
}

func (f githubFile) Truncate(int64) error {
	return ErrReadonly
}

func (f githubFile) WriteString(string) (ret int, err error) {
	return 0, ErrReadonly
}

func (f githubFile) Write([]byte) (n int, err error) {
	return 0, ErrReadonly
}

func (f githubFile) WriteAt([]byte, int64) (n int, err error) {
	return 0, ErrReadonly
}

func (f githubFile) WriteTo(io.Writer) (n int64, err error) {
	return 0, ErrReadonly
}

var _ afero.File = &githubDir{}

type githubDir struct {
	*githubFile

	contents []*github.RepositoryContent
}

func (f *githubDir) Readdir(n int) ([]os.FileInfo, error) {
	// spec: https://pkg.go.dev/os#File.ReadDir

	var ret []*github.RepositoryContent

	// If n <= 0, ReadDir returns all the DirEntry records
	// remaining in the directory. When it succeeds,
	// it returns a nil error (not io.EOF).
	if n <= 0 {
		ret = f.contents
		f.contents = nil
		return repoContentListToFileInfos(ret), nil
	}

	// If n > 0, ReadDir returns at most n DirEntry records.

	possibleN := n
	max := len(f.contents)

	// In this case, if ReadDir returns an empty slice,
	// it will return an error explaining why.
	// At the end of a directory, the error is io.EOF.
	if max == 0 {
		return nil, io.EOF
	}

	// When n > max, it should be truncate to max.
	if possibleN > max {
		possibleN = max // n should always be <= max
	}

	ret = f.contents[:possibleN]
	f.contents = f.contents[possibleN:]
	return repoContentListToFileInfos(ret), nil
}

func (f *githubDir) Readdirnames(n int) ([]string, error) {
	infos, err := f.Readdir(n)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(infos))

	for i, info := range infos {
		names[i] = info.Name()
	}

	return names, nil
}

var _ os.FileInfo = githubFileInfo{}

type githubFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (fi githubFileInfo) Name() string {
	return fi.name
}

func (fi githubFileInfo) Size() int64 {
	return fi.size
}

func (fi githubFileInfo) Mode() os.FileMode {
	return fi.mode
}

func (fi githubFileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi githubFileInfo) IsDir() bool {
	return fi.isDir
}

func (fi githubFileInfo) Sys() any {
	return nil
}

func repoContentToFileInfo(c *github.RepositoryContent) os.FileInfo {
	return &githubFileInfo{
		name:    c.GetName(),
		size:    int64(c.GetSize()),
		modTime: time.Now(), // current time
		mode:    0o444,      // read-only
		isDir:   c.GetType() == "dir",
	}
}

func repoContentListToFileInfos(cs []*github.RepositoryContent) []os.FileInfo {
	ret := make([]os.FileInfo, len(cs))
	for i, c := range cs {
		ret[i] = repoContentToFileInfo(c)
	}
	return ret
}
