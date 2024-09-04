package container

import (
	"context"

	"github.com/thanhhaudev/github-stats/pkg/github"
	"github.com/thanhhaudev/github-stats/pkg/wakatime"
)

type ClientManager struct {
	WakaTimeClient *wakatime.WakaTime
	GitHubClient   *github.GitHub
}

// GetCommits returns the commits of a repository
func (c *ClientManager) GetCommits(ctx context.Context, owner, name, authorID, branch string, numCommits int) ([]github.Commit, error) {
	var allCommits []github.Commit
	var cursor *string

	// Create a new request & set the variables values
	request := github.NewRequest(github.Queries["repository_commits"])
	request.Var("authorId", authorID)
	request.Var("owner", owner)
	request.Var("name", name)
	request.Var("branch", branch)
	request.Var("numCommits", numCommits)

	for {
		if cursor != nil {
			request.Var("afterCursor", *cursor)
		}

		commits, err := c.GitHubClient.Repositories.Commits(ctx, request)
		if err != nil {
			return nil, err
		}

		if commits == nil {
			break
		}

		allCommits = append(allCommits, commits.Nodes...)

		if !commits.PageInfo.HasNextPage {
			break
		}

		cursor = &commits.PageInfo.EndCursor
	}

	return allCommits, nil
}

// GetBranches returns the branches of a repository
func (c *ClientManager) GetBranches(ctx context.Context, owner, name string, numBranches int) ([]github.Branch, error) {
	var allBranches []github.Branch
	var cursor *string
	request := github.NewRequest(github.Queries["repository_branches"])
	request.Var("numBranches", numBranches)
	request.Var("owner", owner)
	request.Var("name", name)

	for {
		if cursor != nil {
			request.Var("afterCursor", *cursor)
		}

		branches, err := c.GitHubClient.Repositories.Branches(ctx, request)
		if err != nil {
			return nil, err
		}

		if branches == nil {
			break
		}

		allBranches = append(allBranches, branches.Nodes...)

		if !branches.PageInfo.HasNextPage {
			break
		}

		cursor = &branches.PageInfo.EndCursor
	}

	return allBranches, nil
}

// GetOwnedRepositories returns the repositories owned or collaborated on by the user
func (c *ClientManager) GetOwnedRepositories(ctx context.Context, username string, numRepos int) ([]github.Repository, error) {
	var allRepos []github.Repository
	var cursor *string
	// Create a new request & set the variables values
	request := github.NewRequest(github.Queries["repositories"])
	request.Var("username", username)
	request.Var("numRepos", numRepos)

	for {
		if cursor != nil {
			request.Var("afterCursor", *cursor)
		}

		repos, err := c.GitHubClient.Repositories.Owned(ctx, request)
		if err != nil {
			return nil, err
		}

		if repos == nil {
			break
		}

		allRepos = append(allRepos, repos.Nodes...)

		if !repos.PageInfo.HasNextPage {
			break
		}

		cursor = &repos.PageInfo.EndCursor
	}

	return allRepos, nil
}

// GetContributedToRepositories returns the repositories contributed to by the user
func (c *ClientManager) GetContributedToRepositories(ctx context.Context, username string, numRepos int) ([]github.Repository, error) {
	var allRepos []github.Repository
	var cursor *string

	// Create a new request & set the variables values
	request := github.NewRequest(github.Queries["repositories_contributed_to"])
	request.Var("username", username)
	request.Var("numRepos", numRepos)

	for {
		if cursor != nil {
			request.Var("afterCursor", *cursor)
		}

		repos, err := c.GitHubClient.Repositories.ContributedTo(ctx, request)
		if err != nil {
			return nil, err
		}

		if repos == nil {
			break
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
func (c *ClientManager) GetViewer(ctx context.Context) (*github.Viewer, error) {
	request := github.NewRequest(github.Queries["viewer"])
	viewer, err := c.GitHubClient.Viewer.Get(ctx, request)
	if err != nil {
		return nil, err
	}

	return viewer, nil
}

// GetDefaultBranch returns the default branch of a repository
func (c *ClientManager) GetDefaultBranch(ctx context.Context, owner, name string) (*github.Branch, error) {
	request := github.NewRequest(github.Queries["repository_default_branch"])
	request.Var("owner", owner)
	request.Var("name", name)

	branch, err := c.GitHubClient.Repositories.DefaultBranch(ctx, request)
	if err != nil {
		return nil, err
	}

	return branch, nil
}

// GetWakaTimeStats returns the user's coding activity statistics
func (c *ClientManager) GetWakaTimeStats(ctx context.Context) (*wakatime.Stats, error) {
	stats, err := c.WakaTimeClient.Stats.Get(ctx)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// NewClientManager creates a new ClientManager
func NewClientManager(w *wakatime.WakaTime, g *github.GitHub) *ClientManager {
	return &ClientManager{w, g}
}
