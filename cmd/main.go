package main

import (
	"context"
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
	logger := log.New(os.Stdout, "", log.Lmsgprefix)
	logger.Println("🚀 Starting...")
	start := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gc := github.NewGitHub(os.Getenv("GITHUB_TOKEN"))
	wc := wakatime.NewWakaTime(os.Getenv("WAKATIME_API_KEY"), wakatime.StatsRange(os.Getenv("WAKATIME_RANGE")))
	dc := container.NewDataContainer(logger, container.NewClientManager(wc, gc))
	if err := dc.Build(ctx); err != nil {
		logger.Fatalln(err)
	}

	sectionName := os.Getenv("SECTION_NAME")
	if sectionName == "" {
		sectionName = "readme-stats"
	}

	cl, err := setClock(logger)
	if err != nil {
		panic(err)
	}

	err = writer.UpdateReadme(dc.GetStats(cl), sectionName)
	if err != nil {
		panic(err)
	}

	logger.Printf("🚩 Execution Duration: %s\n", time.Since(start))
}

func setClock(logger *log.Logger) (clock.Clock, error) {
	cl := clock.NewClock()
	if tz := os.Getenv("TIME_ZONE"); tz != "" {
		err := cl.SetLocation(tz)
		if err != nil {
			logger.Printf("⚠️ Invalid timezone %s: %v\n", tz, err)

			return nil, err
		}

		logger.Printf("🕙 Timezone set to %s\n", tz)
	}

	return cl, nil
}
