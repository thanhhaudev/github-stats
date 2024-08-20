package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/thanhhaudev/github-stats/pkg/writer"
)

func init() {
	godotenv.Load()
}

func main() {
	fmt.Println(getStats())
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
