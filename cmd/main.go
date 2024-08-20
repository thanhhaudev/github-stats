package main

import (
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
	"github.com/thanhhaudev/github-stats/pkg/writer"
)

func main() {
	err := writer.UpdateReadme(getStats(), os.Getenv("SECTION_NAME"))
	if err != nil {
		panic(err)
	}
}

func getStats() string {
	c := NewClientManager(os.Getenv("WAKATIME_API_KEY"), os.Getenv("GITHUB_TOKEN"))
	n, _ := strconv.Atoi(os.Getenv("PER_PAGE"))

	// Get the user's information
	v, err := c.GetViewer()
	if err != nil {
		panic(err)
	}

	// Get the repositories owned by the user
	r, err := c.GetOwnedRepositories(v.Login, n)
	if err != nil {
		panic(err)
	}

	return writer.MakeLanguagePerRepoList(r)
}
