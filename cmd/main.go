package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/thanhhaudev/github-stats/pkg/clock"
	"github.com/thanhhaudev/github-stats/pkg/container"
	"github.com/thanhhaudev/github-stats/pkg/github"
	"github.com/thanhhaudev/github-stats/pkg/wakatime"
	"github.com/thanhhaudev/github-stats/pkg/writer"
)

func main() {
	start := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cl := clock.NewClock()
	if tz := os.Getenv("TIME_ZONE"); tz != "" {
		err := cl.SetLocation(tz)
		if err != nil {
			panic(err)
		}

		fmt.Printf("🕙 Timezone set to %s\n", tz)
	}

	gc := github.NewGitHub(os.Getenv("GITHUB_TOKEN"))
	wc := wakatime.NewWakaTime(os.Getenv("WAKATIME_API_KEY"))
	dc := container.NewDataContainer(container.NewClientManager(wc, gc))
	if gc != nil {
		if err := dc.BuildGitHubData(ctx); err != nil {
			log.Fatalln(err)
		}
	}

	sectionName := os.Getenv("SECTION_NAME")
	if sectionName == "" {
		sectionName = "readme-stats"
	}

	err := writer.UpdateReadme(dc.GetStats(cl), sectionName)
	if err != nil {
		panic(err)
	}

	fmt.Printf("🚀 Execution Duration: %s\n", time.Since(start))
}
