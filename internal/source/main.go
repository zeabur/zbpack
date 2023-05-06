package source

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

type Source interface {
	HasFile(path string) bool
	ReadFile(path string) ([]byte, error)
	ReadDir(path string) ([]FileInfo, error)
}

type FileInfo struct {
	Name  string
	IsDir bool
}

type source struct {
	LocalPath       *string
	GitHubRepoOwner *string
	GitHubRepoName  *string
	GitHubClient    *github.Client
}

func NewLocalSource(path string) Source {
	return &source{LocalPath: &path}
}

func NewGitHubSource(repoOwner, repoName, token string) Source {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	oauth2Client := oauth2.NewClient(context.TODO(), tokenSource)
	client := github.NewClient(oauth2Client)
	return &source{
		GitHubRepoOwner: &repoOwner,
		GitHubRepoName:  &repoName,
		GitHubClient:    client,
	}
}

func (s source) HasFile(path string) bool {
	if s.LocalPath != nil {
		_, err := os.Stat(*s.LocalPath + "/" + path)
		return err == nil
	}
	if s.GitHubRepoOwner != nil && s.GitHubRepoName != nil {
		_, _, _, err := s.GitHubClient.Repositories.GetContents(
			context.Background(), *s.GitHubRepoOwner, *s.GitHubRepoName, path,
			nil,
		)
		return err == nil
	}
	return false
}

func (s source) ReadFile(path string) ([]byte, error) {
	if s.LocalPath != nil {
		return os.ReadFile(*s.LocalPath + "/" + path)
	}

	if s.GitHubRepoOwner != nil && s.GitHubRepoName != nil {
		file, _, _, err := s.GitHubClient.Repositories.GetContents(
			context.Background(), *s.GitHubRepoOwner, *s.GitHubRepoName, path,
			nil,
		)
		if err != nil {
			return nil, err
		}
		c, err := file.GetContent()
		if err != nil {
			return nil, err
		}
		return []byte(c), nil
	}

	return nil, fmt.Errorf("no source specified")
}

func (s source) ReadDir(path string) ([]FileInfo, error) {
	if s.LocalPath != nil {
		dir, err := os.ReadDir(*s.LocalPath + "/" + path)
		if err != nil {
			return nil, err
		}
		var res []FileInfo
		for _, f := range dir {
			res = append(
				res, FileInfo{
					Name:  f.Name(),
					IsDir: f.IsDir(),
				},
			)
		}
		return res, nil
	}
	if s.GitHubRepoOwner != nil && s.GitHubRepoName != nil {
		_, files, _, err := s.GitHubClient.Repositories.GetContents(
			context.Background(), *s.GitHubRepoOwner, *s.GitHubRepoName, path,
			nil,
		)
		if err != nil {
			return nil, err
		}
		var res []FileInfo
		for _, f := range files {
			res = append(
				res, FileInfo{
					Name:  f.GetName(),
					IsDir: f.GetType() == "dir",
				},
			)
		}
		return res, nil
	}
	return nil, fmt.Errorf("no source specified")
}
