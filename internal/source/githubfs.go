package source

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/go-github/v63/github"
	"github.com/spf13/afero"
	"github.com/spf13/afero/zipfs"
)

type githubFsOptions struct {
	owner string
	name  string
	ref   string
}

// GitHubFsOption is the option for NewGitHubFs.
type GitHubFsOption func(*githubFsOptions)

// GitHubRef sets the ref of the GitHub repository.
func GitHubRef(ref string) GitHubFsOption {
	return func(fs *githubFsOptions) {
		fs.ref = ref
	}
}

// NewGitHubFs creates a new github filesystem.
func NewGitHubFs(repoOwner, repoName string, token *string, options ...GitHubFsOption) (afero.Fs, error) {
	fsOptions := &githubFsOptions{
		owner: repoOwner,
		name:  repoName,
	}

	for _, opt := range options {
		opt(fsOptions)
	}

	fs, err := getRefZipFs(fsOptions.owner, fsOptions.name, fsOptions.ref, token)
	if err != nil {
		return nil, fmt.Errorf("get ref tarball fs: %w", err)
	}

	return fs, nil
}

const zipFileSizeLimit = 1024 * 1024 * 1024 /* 1 GiB */

func getRefZipFs(owner, name, ref string, token *string) (afero.Fs, error) {
	githubClient := github.NewClient(nil)
	if token != nil {
		githubClient = githubClient.WithAuthToken(*token)
	}

	repo, _, err := githubClient.Repositories.GetArchiveLink(context.Background(), owner, name, github.Zipball, &github.RepositoryContentGetOptions{
		Ref: ref,
	}, 1)
	if err != nil {
		return nil, fmt.Errorf("get archive link: %w", err)
	}

	content, err := githubClient.Client().Get(repo.String())
	if err != nil {
		return nil, fmt.Errorf("get tarball: %w", err)
	}
	defer func() {
		_ = content.Body.Close()
	}()

	b, n, err := ReadLimited(content.Body, zipFileSizeLimit+1)
	if err != nil {
		return nil, fmt.Errorf("read from GitHub response: %w", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(b), n)
	if err != nil {
		return nil, fmt.Errorf("new zip reader: %w", err)
	}

	fs := zipfs.New(zipReader)

	// A GitHub zipball contains a single directory that includes the repository root.
	// Since the directory name is not fixed or deterministic, we need to find it.
	directories, err := afero.ReadDir(fs, "")
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}
	if len(directories) == 1 && directories[0].IsDir() {
		return afero.NewBasePathFs(fs, directories[0].Name()), nil
	}

	return fs, nil
}

// ErrOverSized is the error when the reader exceeds the limit.
var ErrOverSized = errors.New("oversized reader")

// ReadLimited reads from r until limit bytes or EOF, whichever comes first.
// If the limit is exceeded, it returns an error.
//
// It returns the content read, the number of bytes read, and any error occurred.
func ReadLimited(r io.Reader, limit int) (content []byte, n int64, err error) {
	buf := new(bytes.Buffer)
	n, err = buf.ReadFrom(io.LimitReader(r, int64(limit)+1))
	if err != nil {
		return nil, n, err
	}
	if n > int64(limit) {
		return nil, n, ErrOverSized
	}
	return buf.Bytes(), n, nil
}
