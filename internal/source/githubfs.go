package source

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

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

	zipballStreamReader := io.LimitReader(content.Body, zipFileSizeLimit+1)

	buf := new(bytes.Buffer)
	n, err := buf.ReadFrom(zipballStreamReader)
	if err != nil {
		return nil, fmt.Errorf("read tarball: %w", err)
	}
	if n > zipFileSizeLimit {
		return nil, fmt.Errorf("repo is too large; limit is %f GiB", float64(zipFileSizeLimit)/1024/1024)
	}

	zipballReadAtReader := bytes.NewReader(buf.Bytes())

	zipReader, err := zip.NewReader(zipballReadAtReader, n)
	if err != nil {
		return nil, fmt.Errorf("new zip reader: %w", err)
	}

	fs := zipfs.New(zipReader)

	filename := content.Header.Get("Content-Disposition")
	if attachmentName, ok := strings.CutPrefix(filename, "attachment; filename="); ok {
		return afero.NewBasePathFs(fs, strings.TrimSuffix(attachmentName, ".zip")), nil
	}

	return fs, nil
}
