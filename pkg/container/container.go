package container

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/thanhhaudev/github-stats/pkg/clock"
	"github.com/thanhhaudev/github-stats/pkg/github"
	"github.com/thanhhaudev/github-stats/pkg/wakatime"
	"github.com/thanhhaudev/github-stats/pkg/writer"
)

const (
	repoPerQuery   = 25
	branchPerQuery = 30
	commitPerQuery = 100
)

type DataContainer struct {
	ClientManager *ClientManager
	Logger        *log.Logger
	Data          struct {
		Viewer       *github.Viewer
		Repositories []github.Repository
		Commits      []github.Commit
		WakaTime     *wakatime.Stats
	}
}

// metrics returns the metrics map
func (d *DataContainer) metrics(com *CommitStats, lang *LanguageStats) map[string]string {
	return map[string]string{
		"LANGUAGE_PER_REPO":   writer.MakeLanguagePerRepoList(d.Data.Repositories),
		"LANGUAGES_AND_TOOLS": writer.MakeLanguageAndToolList(lang.Languages, lang.TotalSize),
		"COMMIT_DAYS_OF_WEEK": writer.MakeCommitDaysOfWeekList(com.DailyCommits, com.TotalCommits),
		"COMMIT_TIME_OF_DAY":  writer.MakeCommitTimeOfDayList(d.Data.Commits),
		"WAKATIME_SPENT_TIME": writer.MakeWakaActivityList(
			d.Data.WakaTime,
			strings.Split(os.Getenv("WAKATIME_DATA"), ","),
		),
	}
}

// GetStats returns the statistics
func (d *DataContainer) GetStats(cl clock.Clock) string {
	d.Logger.Println("Creating statistics...")
	b := strings.Builder{}

	// show metrics based on the environment variable
	w := d.metrics(d.CalculateCommits(), d.CalculateLanguages())
	for _, k := range strings.Split(os.Getenv("SHOW_METRICS"), ",") {
		v, ok := w[k]
		if !ok {
			continue
		}

		b.WriteString(v)
	}

	// Show last update time if enabled
	showLastUpdated(cl, &b)

	d.Logger.Println("Created statistics successfully")

	return b.String()
}

func showLastUpdated(cl clock.Clock, b *strings.Builder) {
	if os.Getenv("SHOW_LAST_UPDATE") == "true" {
		layout := os.Getenv("TIME_LAYOUT")
		if layout == "" {
			layout = clock.DateTimeFormatWithTimezone
		}

		b.WriteString(writer.MakeLastUpdatedOn(cl.Now().Format(layout)))
	}
}

// InitViewer initializes the viewer
func (d *DataContainer) InitViewer(ctx context.Context) error {
	d.Logger.Println("Fetching viewer information...")

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
	d.Logger.Println("Fetching repositories...")
	seenRepos := make(map[string]bool)
	errChan := make(chan error, 2)
	repoChan := make(chan []github.Repository, 2)
	isExcludeForks := os.Getenv("EXCLUDE_FORK_REPOS") == "true"

	go func() {
		r, err := d.ClientManager.GetOwnedRepositories(ctx, d.Data.Viewer.Login, repoPerQuery)
		if err != nil {
			errChan <- err
			return
		}

		repoChan <- r
		errChan <- nil

		d.Logger.Println("Fetched owned repositories successfully")
	}()

	go func() {
		c, err := d.ClientManager.GetContributedToRepositories(ctx, d.Data.Viewer.Login, repoPerQuery)
		if err != nil {
			errChan <- err
			return
		}

		repoChan <- c
		errChan <- nil

		d.Logger.Println("Fetched contributed to repositories successfully")
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
			if isExcludeForks && repo.IsFork {
				continue
			}

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
	d.Logger.Println("Fetching commits...")
	fetchAllBranches := os.Getenv("ONLY_MAIN_BRANCH") != "true"
	hiddenRepoInfo := os.Getenv("HIDE_REPO_INFO") == "true"
	repoCount := len(d.Data.Repositories)
	errChan := make(chan error, repoCount)
	commitChan := make(chan []github.Commit, repoCount)
	seenOIDs := make(map[string]bool)
	mask := func(input string) string {
		length := len(input)
		if length <= 2 {
			return input // No masking for very short strings
		}

		num := length / 3
		prefixLength := (length - num) / 2
		suffixLength := length - prefixLength - num

		return input[:prefixLength] + strings.Repeat("*", num) + input[len(input)-suffixLength:]
	}

	if hiddenRepoInfo {
		d.Logger.Printf("Fetching commits from %d %s...", repoCount, func() string {
			if repoCount == 1 {
				return "repository"
			}

			return "repositories"
		}())
	}

	for _, repo := range d.Data.Repositories {
		go func(repo github.Repository) {
			if fetchAllBranches {
				if !hiddenRepoInfo {
					d.Logger.Println("Fetching commits from all branches of repository:", mask(repo.Name))
				}

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
				if !hiddenRepoInfo {
					d.Logger.Println("Fetching commits from the default branch of repository:", mask(repo.Name))
				}

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
				commit.CommittedDate = ctx.Value(clock.ClockKey{}).(clock.Clock).ToClockTz(commit.CommittedDate)
				d.Data.Commits = append(d.Data.Commits, commit)
			}
		}
	}

	d.Logger.Println("Fetched commits successfully")

	return nil
}

// InitWakaStats initializes the WakaTime statistics
func (d *DataContainer) InitWakaStats(ctx context.Context) error {
	d.Logger.Println("Fetching WakaTime statistics...")

	v, err := d.ClientManager.GetWakaTimeStats(ctx)
	if err != nil {
		return err
	}

	// if the status is not "ok", print the message
	if v.Data.Status != "ok" {
		switch v.Data.Status {
		case "pending_update":
			d.Logger.Println("WakaTime is not ready yet")
		default:
			d.Logger.Println("An error occurred while fetching WakaTime data:", v.Data.Status)

			return nil // Skip if the status is unknown
		}
	}

	d.Data.WakaTime = v

	return nil
}

// Build builds the data container
func (d *DataContainer) Build(ctx context.Context) error {
	d.Logger.Println("Building data container...")

	// if the GitHub client is not nil, initialize the viewer, repositories, and commits
	if d.ClientManager.GitHubClient != nil {
		d.Logger.Println("Fetching data from GitHub APIs...")
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

		d.Logger.Println("Fetching data from GitHub APIs successfully")
	}

	// if the WakaTime client is not nil, fetch data from WakaTime APIs
	if d.ClientManager.WakaTimeClient != nil {
		d.Logger.Println("Fetching data from Wakatime APIs...")
		err := d.InitWakaStats(ctx)
		if err != nil {
			return err
		}

		d.Logger.Println("Fetching data from Wakatime APIs successfully")
	}

	d.Logger.Println("Built data container successfully")

	return nil
}

// NewDataContainer creates a new DataContainer
func NewDataContainer(l *log.Logger, cm *ClientManager) *DataContainer {
	return &DataContainer{
		Logger:        l,
		ClientManager: cm,
	}
}
