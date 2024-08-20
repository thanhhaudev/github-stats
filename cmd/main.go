package main

import (
	"fmt"
	"os"

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

	r, err := c.GetOwnedRepositories(os.Getenv("GITHUB_USERNAME"), 1)
	if err != nil {
		panic(err)
	}

	return writer.MakeLanguagePerRepoList(r)
}
