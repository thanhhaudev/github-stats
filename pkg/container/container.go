package container

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/thanhhaudev/github-stats/pkg/clock"
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

// GetGithubWidgets returns the widgets to display
func (d *DataContainer) GetGithubWidgets(com *CommitStats) map[string]string {
	return map[string]string{
		"LANGUAGE_PER_REPO":   writer.MakeLanguagePerRepoList(d.Data.Repositories),
		"COMMIT_DAYS_OF_WEEK": writer.MakeCommitDaysOfWeekList(com.DailyCommits, com.TotalCommits),
		"COMMIT_TIME_OF_DAY":  writer.MakeCommitTimeOfDayList(d.Data.Commits),
	}
}

// GetStats returns the statistics
func (d *DataContainer) GetStats(cl clock.Clock) string {
	fmt.Println("Calculating statistics...")
	b := strings.Builder{}

	// show GitHub widgets if enabled
	if d.ClientManager.GitHubClient != nil {
		fmt.Println("Creating GitHub widgets...")
		w := d.GetGithubWidgets(d.CalculateCommits(cl))
		showGitHubWidgets(w, &b)
	}

	// Show last update time if enabled
	showLastUpdated(cl, &b)

	fmt.Println("Calculated statistics successfully")

	return b.String()
}

func showGitHubWidgets(w map[string]string, b *strings.Builder) {
	for _, k := range strings.Split(os.Getenv("GITHUB_WIDGETS"), ",") {
		v, ok := w[k]
		if !ok {
			continue
		}

		b.WriteString(v)
	}
}

func showLastUpdated(cl clock.Clock, b *strings.Builder) {
	if os.Getenv("SHOW_LAST_UPDATE") == "true" {
		layout := os.Getenv("TIME_LAYOUT")
		if layout == "" {
			layout = clock.DateTimeFormatWithTimezone
		} else {
			fmt.Println("Using custom time layout:", layout)
		}

		b.WriteString(writer.MakeLastUpdatedOn(cl.Now().Format(layout)))
	}
}

// InitViewer initializes the viewer
func (d *DataContainer) InitViewer(ctx context.Context) error {
	fmt.Println("Fetching viewer information...")

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
	fmt.Println("Fetching repositories...")
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

		fmt.Println("Fetched owned repositories successfully")
	}()

	go func() {
		c, err := d.ClientManager.GetContributedToRepositories(ctx, d.Data.Viewer.Login, repoPerQuery)
		if err != nil {
			errChan <- err
			return
		}

		repoChan <- c
		errChan <- nil

		fmt.Println("Fetched contributed to repositories successfully")
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
	fmt.Println("Fetching commits...")
	fetchAllBranches := os.Getenv("ONLY_MAIN_BRANCH") != "true"
	repoCount := len(d.Data.Repositories)
	errChan := make(chan error, repoCount)
	commitChan := make(chan []github.Commit, repoCount)
	seenOIDs := make(map[string]bool)

	for _, repo := range d.Data.Repositories {
		go func(repo github.Repository) {
			if fetchAllBranches {
				fmt.Println("Fetching commits from all branches of repository:", repo.Name)
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
			} else {
				fmt.Println("Fetching commits from the default branch of repository:", repo.Name)
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

	fmt.Println("Fetched commits successfully")

	return nil
}

// BuildGitHubData builds the data container
func (d *DataContainer) BuildGitHubData(ctx context.Context) error {
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
func NewDataContainer(cm *ClientManager) *DataContainer {
	return &DataContainer{
		ClientManager: cm,
	}
}
