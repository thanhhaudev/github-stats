package main

import (
	"os"
	"strings"
	"time"

	"github.com/thanhhaudev/github-stats/pkg/github"
	"github.com/thanhhaudev/github-stats/pkg/writer"
)

type DataContainer struct {
	ClientManager *ClientManager
	Data          struct {
		Viewer            *github.Viewer
		OwnedRepositories []github.Repository
	}
}

func (d *DataContainer) GetWidgets() map[string]string {
	return map[string]string{
		"LANGUAGE_PER_REPO": writer.MakeLanguagePerRepoList(d.Data.OwnedRepositories),
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

func (d *DataContainer) InitOwnedRepositories() {
	r, err := d.ClientManager.GetOwnedRepositories(d.Data.Viewer.Login, 25)
	if err != nil {
		panic(err)
	}

	d.Data.OwnedRepositories = r
}

func (d *DataContainer) Build() {
	d.InitViewer()
	d.InitOwnedRepositories()

	time.Sleep(time.Second)
}

// NewDataContainer creates a new DataContainer
func NewDataContainer() *DataContainer {
	c := NewClientManager(os.Getenv("WAKATIME_API_KEY"), os.Getenv("GITHUB_TOKEN"))
	return &DataContainer{
		ClientManager: c,
	}
}
