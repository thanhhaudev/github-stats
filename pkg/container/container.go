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
	Context       context.Context
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
func (d *DataContainer) InitViewer() {
	v, err := d.ClientManager.GetViewer(d.Context)
	if err != nil {
		panic(err)
	}

	d.Data.Viewer = v
}

// InitRepositories initializes the repositories
// owned and contributed to by the user
func (d *DataContainer) InitRepositories() {
	r, err := d.ClientManager.GetOwnedRepositories(d.Context, d.Data.Viewer.Login, repoPerQuery)
	if err != nil {
		panic(err)
	}

	// Get the unique URLs of the repositories
	u := make(map[string]bool)
	for _, repo := range r {
		u[repo.Url] = true
	}

	c, err := d.ClientManager.GetContributedToRepositories(d.Context, d.Data.Viewer.Login, repoPerQuery)
	if err != nil {
		panic(err)
	}

	for _, repo := range c {
		if _, ok := u[repo.Url]; !ok { // Only add the repository if it is not already in the list
			r = append(r, repo)
		}
	}

	d.Data.Repositories = r
}

// InitCommits initializes the branches of the repositories
func (d *DataContainer) InitCommits() {
	for _, repo := range d.Data.Repositories {
		b, err := d.ClientManager.GetBranches(d.Context, repo.Owner.Login, repo.Name, branchPerQuery)
		if err != nil {
			panic(err)
		}

		for _, branch := range b {
			commits, err := d.ClientManager.GetCommits(d.Context, repo.Owner.Login, repo.Name, d.Data.Viewer.ID, fmt.Sprintf("refs/heads/%s", branch.Name), commitPerQuery)
			if err != nil {
				panic(err)
			}

			d.Data.Commits = append(d.Data.Commits, commits...)
		}
	}
}

// Build builds the data container
func (d *DataContainer) Build() {
	d.InitViewer()
	d.InitRepositories()
	d.InitCommits()
}

// NewDataContainer creates a new DataContainer
func NewDataContainer(ctx context.Context) *DataContainer {
	return &DataContainer{
		ClientManager: NewClientManager(os.Getenv("WAKATIME_API_KEY"), os.Getenv("GITHUB_TOKEN")),
		Context:       ctx,
	}
}
