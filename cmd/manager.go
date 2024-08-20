package main

import (
	"github.com/thanhhaudev/github-stats/pkg/github"
	"github.com/thanhhaudev/github-stats/pkg/wakatime"
)

type ClientManager struct {
	WakaTimeClient *wakatime.WakaTime
	GitHubClient   *github.GitHub
}

// GetOwnedRepositories returns the repositories owned by the user
func (c *ClientManager) GetOwnedRepositories(username string, numRepos int) ([]github.Repository, error) {
	var allRepos []github.Repository
	var cursor *string

	for {
		request := github.NewRequest(github.Queries["repositories"])
		request.Var("username", username)
		request.Var("numRepos", numRepos)
		if cursor != nil {
			request.Var("afterCursor", *cursor)
		}

		repos, err := c.GitHubClient.Repositories.Owned(request)
		if err != nil {
			return nil, err
		}

		allRepos = append(allRepos, repos.Nodes...)

		if !repos.PageInfo.HasNextPage {
			break
		}

		cursor = &repos.PageInfo.EndCursor
	}

	return allRepos, nil
}

// GetViewer returns the viewer's information
func (c *ClientManager) GetViewer() (*github.Viewer, error) {
	request := github.NewRequest(github.Queries["viewer"])
	viewer, err := c.GitHubClient.Viewer.Get(request)
	if err != nil {
		return nil, err
	}

	return viewer, nil
}

// NewClientManager creates a new ClientManager
func NewClientManager(wakaTimeApiKey, gitHubApiKey string) *ClientManager {
	return &ClientManager{
		WakaTimeClient: wakatime.NewWakaTime(wakaTimeApiKey),
		GitHubClient:   github.NewGitHub(gitHubApiKey),
	}
}
