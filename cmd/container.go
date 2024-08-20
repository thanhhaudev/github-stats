package main

import (
	"os"
	"strings"
	"time"

	"github.com/thanhhaudev/github-stats/pkg/github"
	"github.com/thanhhaudev/github-stats/pkg/writer"
)

const repoPerPage = 25

type DataContainer struct {
	ClientManager *ClientManager
	Data          struct {
		Viewer       *github.Viewer
		Repositories []github.Repository
	}
}

func (d *DataContainer) GetWidgets() map[string]string {
	return map[string]string{
		"LANGUAGE_PER_REPO": writer.MakeLanguagePerRepoList(d.Data.Repositories),
	}
}

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

func (d *DataContainer) InitViewer() {
	v, err := d.ClientManager.GetViewer()
	if err != nil {
		panic(err)
	}

	d.Data.Viewer = v
}

// InitRepositories initializes the repositories
// owned and contributed to by the user
func (d *DataContainer) InitRepositories() {
	r, err := d.ClientManager.GetOwnedRepositories(d.Data.Viewer.Login, repoPerPage)
	if err != nil {
		panic(err)
	}

	// Get the unique URLs of the repositories
	u := make(map[string]bool)
	for _, repo := range r {
		u[repo.Url] = true
	}

	c, err := d.ClientManager.GetContributedToRepositories(d.Data.Viewer.Login, repoPerPage)
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

func (d *DataContainer) Build() {
	d.InitViewer()
	d.InitRepositories()

	time.Sleep(time.Second)
}

// NewDataContainer creates a new DataContainer
func NewDataContainer() *DataContainer {
	c := NewClientManager(os.Getenv("WAKATIME_API_KEY"), os.Getenv("GITHUB_TOKEN"))
	return &DataContainer{
		ClientManager: c,
	}
}
