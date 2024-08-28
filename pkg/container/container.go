package container

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/thanhhaudev/github-stats/pkg/github"
	"github.com/thanhhaudev/github-stats/pkg/writer"
)

const (
	repoPerQuery   = 25
	branchPerQuery = 30
	commitPerQuery = 100
)

type DataContainer struct {
	ClientManager *ClientManager
	Data          struct {
		Viewer       *github.Viewer
		Repositories []github.Repository
		Commits      []github.Commit
	}
}

// GetWidgets returns the widgets to display
func (d *DataContainer) GetWidgets() map[string]string {
	data := d.CalculateCommits()

	return map[string]string{
		"LANGUAGE_PER_REPO":   writer.MakeLanguagePerRepoList(d.Data.Repositories),
		"COMMIT_DAYS_OF_WEEK": writer.MakeCommitDaysOfWeekList(data.DailyCommits, data.TotalCommits),
		"COMMIT_TIME_OF_DAY":  writer.MakeCommitTimeOfDayList(d.Data.Commits),
	}
}

// GetStats returns the statistics
func (d *DataContainer) GetStats() string {
	b := strings.Builder{}
	w := d.GetWidgets()
	for _, k := range strings.Split(os.Getenv("SHOW_WIDGETS"), ",") {
		v, ok := w[k]
		if !ok {
			continue
		}

		b.WriteString(v)
	}

	return b.String()
}

// InitViewer initializes the viewer
func (d *DataContainer) InitViewer(ctx context.Context) error {
	v, err := d.ClientManager.GetViewer(ctx)
	if err != nil {
		return err
	}

	d.Data.Viewer = v

	return nil
}

// InitRepositories initializes the repositories
// owned and contributed to by the user
func (d *DataContainer) InitRepositories(ctx context.Context) error {
	seenRepos := make(map[string]bool)
	errChan := make(chan error, 2)
	repoChan := make(chan []github.Repository, 2)

	go func() {
		r, err := d.ClientManager.GetOwnedRepositories(ctx, d.Data.Viewer.Login, repoPerQuery)
		if err != nil {
			errChan <- err
			return
		}

		repoChan <- r
		errChan <- nil
	}()

	go func() {
		c, err := d.ClientManager.GetContributedToRepositories(ctx, d.Data.Viewer.Login, repoPerQuery)
		if err != nil {
			errChan <- err
			return
		}

		repoChan <- c
		errChan <- nil
	}()

	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}

	close(repoChan) // Close the channel to signal that all repositories have been fetched

	// Deduplicate repositories
	for repos := range repoChan {
		for _, repo := range repos {
			if !seenRepos[repo.Url] {
				seenRepos[repo.Url] = true
				d.Data.Repositories = append(d.Data.Repositories, repo)
			}
		}
	}

	return nil
}

// InitCommits initializes the branches of the repositories
func (d *DataContainer) InitCommits(ctx context.Context) error {
	fetchAllBranches := os.Getenv("ONLY_MAIN_BRANCH") != "true"
	repoCount := len(d.Data.Repositories)
	errChan := make(chan error, repoCount)
	commitChan := make(chan []github.Commit, repoCount)
	seenOIDs := make(map[string]bool)

	for _, repo := range d.Data.Repositories {
		go func(repo github.Repository) {
			if fetchAllBranches { // Fetch commits from all branches
				branches, err := d.ClientManager.GetBranches(ctx, repo.Owner.Login, repo.Name, branchPerQuery)
				if err != nil {
					errChan <- err
					return
				}

				var allCommits []github.Commit
				for _, branch := range branches {
					commits, err := d.ClientManager.GetCommits(ctx, repo.Owner.Login, repo.Name, d.Data.Viewer.ID, fmt.Sprintf("refs/heads/%s", branch.Name), commitPerQuery)
					if err != nil {
						errChan <- err
						return
					}

					allCommits = append(allCommits, commits...)
				}

				commitChan <- allCommits
			} else { // Fetch commits from the default branch
				defaultBranch, err := d.ClientManager.GetDefaultBranch(ctx, repo.Owner.Login, repo.Name)
				if err != nil {
					errChan <- err
					return
				}

				commits, err := d.ClientManager.GetCommits(ctx, repo.Owner.Login, repo.Name, d.Data.Viewer.ID, fmt.Sprintf("refs/heads/%s", defaultBranch.Name), commitPerQuery)
				if err != nil {
					errChan <- err
					return
				}

				commitChan <- commits
			}
			errChan <- nil
		}(repo)
	}

	for i := 0; i < len(d.Data.Repositories); i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}

	close(commitChan) // Close the channel to signal that all commits have been fetched

	// Deduplicate commits
	for commits := range commitChan {
		for _, commit := range commits {
			if !seenOIDs[commit.OID] {
				seenOIDs[commit.OID] = true
				d.Data.Commits = append(d.Data.Commits, commit)
			}
		}
	}

	return nil
}

// Build builds the data container
func (d *DataContainer) Build(ctx context.Context) error {
	err := d.InitViewer(ctx)
	if err != nil {
		return err
	}

	err = d.InitRepositories(ctx)
	if err != nil {
		return err
	}

	err = d.InitCommits(ctx)
	if err != nil {
		return err
	}

	return nil
}

// NewDataContainer creates a new DataContainer
func NewDataContainer() *DataContainer {
	return &DataContainer{
		ClientManager: NewClientManager(os.Getenv("WAKATIME_API_KEY"), os.Getenv("GITHUB_TOKEN")),
	}
}
