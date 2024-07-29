package source

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/sideband"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/spf13/afero"
)

var _ afero.Fs = githubFs{}

type githubFs struct {
	owner string
	name  string
	ref   string

	worktree *git.Worktree
}

// GitHubFsOption is the option for NewGitHubFs.
type GitHubFsOption func(*githubFs)

// GitHubRef sets the ref of the GitHub repository.
func GitHubRef(ref string) GitHubFsOption {
	return func(fs *githubFs) {
		fs.ref = ref
	}
}

// NewGitHubFs creates a new github filesystem.
func NewGitHubFs(repoOwner, repoName, token string, options ...GitHubFsOption) (afero.Fs, error) {
	fs := &githubFs{
		owner: repoOwner,
		name:  repoName,
	}

	for _, opt := range options {
		opt(fs)
	}

	muxer := sideband.NewMuxer(sideband.Sideband64k, os.Stderr)

	r, err := git.Clone(memory.NewStorage(), memfs.New(), &git.CloneOptions{
		URL: fmt.Sprintf("https://github.com/%s/%s", fs.owner, fs.name),
		Auth: &http.BasicAuth{
			Username: "",
			Password: token,
		},
		NoCheckout:        fs.ref != "", // if ref is given, we checkout later
		RecurseSubmodules: 1,
		ShallowSubmodules: true,
		Progress:          muxer,
		Mirror:            true,
	})
	if err != nil {
		return nil, fmt.Errorf("clone repository: %w", err)
	}

	wt, err := r.Worktree()
	if err != nil {
		return nil, fmt.Errorf("retrieve worktree: %w", err)
	}

	if fs.ref != "" {
		// checkout to the given ref

		hash, err := r.ResolveRevision(plumbing.Revision(fs.ref))
		if err != nil {
			return nil, fmt.Errorf("resolve revision: %w", err)
		}

		err = wt.Checkout(&git.CheckoutOptions{
			Hash: *hash,
		})
		if err != nil {
			return nil, fmt.Errorf("checkout: %w", err)
		}
	}

	fs.worktree = wt

	return fs, nil
}

var (
	// ErrUnimplemented is returned when a method is not implemented.
	ErrUnimplemented = errors.New("unimplemented")
	// ErrReadonly is returned when this filesystem is readonly.
	ErrReadonly = errors.New("readonly")
	// ErrNotDir is returned when something is not a directory.
	ErrNotDir = errors.New("not a directory")
	// ErrNotFile is returned when something is not a file.
	ErrNotFile = errors.New("not a file")
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
	fstat, err := fs.Stat(name)
	if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}

	if fstat.IsDir() {
		dirFiles, err := fs.worktree.Filesystem.ReadDir(name)
		if err != nil {
			return nil, fmt.Errorf("read dir: %w", err)
		}

		return &githubDir{
			dirFiles: dirFiles,
			info:     fstat,
		}, nil
	}

	file, err := fs.worktree.Filesystem.Open(name)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	return &githubFile{
		File: file,
		info: nil,
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
	return fs.worktree.Filesystem.Stat(name)
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
	billy.File
	info os.FileInfo
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
	return nil
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
	name     string
	dirFiles []os.FileInfo
	info     os.FileInfo
}

func (f *githubDir) Close() error {
	return nil
}

func (f *githubDir) Name() string {
	return f.name
}

func (f *githubDir) Read([]byte) (n int, err error) {
	return 0, ErrNotFile
}

func (f *githubDir) ReadAt([]byte, int64) (n int, err error) {
	return 0, ErrNotFile
}

func (f *githubDir) Seek(int64, int) (int64, error) {
	return 0, ErrNotFile
}

func (f *githubDir) Stat() (fs.FileInfo, error) {
	return f.info, nil
}

func (f *githubDir) Sync() error {
	return nil
}

func (f *githubDir) Truncate(int64) error {
	return ErrNotFile
}

func (f *githubDir) Write([]byte) (n int, err error) {
	return 0, ErrNotFile
}

func (f *githubDir) WriteAt([]byte, int64) (n int, err error) {
	return 0, ErrNotFile
}

func (f *githubDir) WriteString(string) (ret int, err error) {
	return 0, ErrNotFile
}

func (f *githubDir) Readdir(n int) ([]os.FileInfo, error) {
	// spec: https://pkg.go.dev/os#File.ReadDir
	// If n <= 0, ReadDir returns all the DirEntry records
	// remaining in the directory. When it succeeds,
	// it returns a nil error (not io.EOF).
	if n <= 0 {
		return f.dirFiles, nil
	}

	// If n > 0, ReadDir returns at most n DirEntry records.

	entriesToReturn := min(n, len(f.dirFiles))
	if entriesToReturn == 0 {
		return nil, io.EOF
	}

	ret := f.dirFiles[:entriesToReturn]
	return ret, nil
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
