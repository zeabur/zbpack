package source

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v63/github"
	"github.com/spf13/afero"
	"github.com/spf13/afero/tarfs"
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
func NewGitHubFs(repoOwner, repoName, token string, options ...GitHubFsOption) (afero.Fs, error) {
	fsOptions := &githubFsOptions{
		owner: repoOwner,
		name:  repoName,
	}

	for _, opt := range options {
		opt(fsOptions)
	}

	fs, err := getRefTarballFs(fsOptions.owner, fsOptions.name, fsOptions.ref, token)
	if err != nil {
		return nil, fmt.Errorf("get ref tarball fs: %w", err)
	}

	return fs, nil
}

func getRefTarballFs(owner, name, ref string, token string) (afero.Fs, error) {
	githubClient := github.NewClient(nil).WithAuthToken(token)

	repo, _, err := githubClient.Repositories.GetArchiveLink(context.Background(), owner, name, github.Tarball, &github.RepositoryContentGetOptions{
		Ref: ref,
	}, 1)
	if err != nil {
		return nil, fmt.Errorf("get archive link: %w", err)
	}

	tarContent, err := githubClient.Client().Get(repo.String())
	if err != nil {
		return nil, fmt.Errorf("get tarball: %w", err)
	}
	defer func() {
		_ = tarContent.Body.Close()
	}()

	// move tar content to memory
	buffer := &bytes.Buffer{}
	if _, err := buffer.ReadFrom(tarContent.Body); err != nil {
		return nil, fmt.Errorf("read from tar content: %w", err)
	}

	gzipReader, err := gzip.NewReader(buffer)
	if err != nil {
		return nil, fmt.Errorf("new gzip reader: %w", err)
	}

	tarReader := tar.NewReader(gzipReader)
	fs := tarfs.New(tarReader)

	filename := tarContent.Header.Get("Content-Disposition")
	if attachmentName, ok := strings.CutPrefix(filename, "attachment; filename="); ok {
		return afero.NewBasePathFs(fs, strings.TrimSuffix(attachmentName, ".tar.gz")), nil
	}

	return fs, nil
}
