package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/thanhhaudev/github-stats/pkg/clock"
	"github.com/thanhhaudev/github-stats/pkg/container"
	"github.com/thanhhaudev/github-stats/pkg/github"
	"github.com/thanhhaudev/github-stats/pkg/wakatime"
)

func main() {
	logger := log.New(os.Stdout, "", log.Lmsgprefix)
	logger.Println("🚀 Starting...")
	cl, err := setClock(logger)
	if err != nil {
		panic(err)
	}

	start := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	ctx = withClock(ctx, cl)
	defer cancel()

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		logger.Fatalln("❌ GITHUB_TOKEN is required for this action")
	}

	gc := github.NewGitHub(token)
	wc := wakatime.NewWakaTime(logger, os.Getenv("WAKATIME_API_KEY"), wakatime.StatsRange(os.Getenv("WAKATIME_RANGE")))
	dc := container.NewDataContainer(logger, container.NewClientManager(wc, gc))
	if err := dc.Build(ctx); err != nil {
		logger.Fatalln(err)
	}

	logger.Println("📝 Updating README.md...")
	err = updateReadme(dc.GetStats(cl), os.Getenv("SECTION_NAME"))
	if err != nil {
		logger.Fatalf("Error updating README.md: %v", err)
	}

	if os.Getenv("DRY_RUN") != "true" {
		logger.Println("🔧 Setting up git config...")
		err = setupGitConfig(
			dc.Data.Viewer.Login,
			os.Getenv("GITHUB_TOKEN"),
			os.Getenv("COMMIT_USER_NAME"),
			os.Getenv("COMMIT_USER_EMAIL"),
		)
		if err != nil {
			logger.Fatalf("Error setting up git config: %v", err)
		}

		changed, err := hasReadmeChanged()
		if err != nil {
			logger.Fatalf("Error checking if README.md has changed: %v", err)
		}

		if changed {
			logger.Println("📤 Committing and pushing changes...")
			err = commitAndPushReadme(os.Getenv("COMMIT_MESSAGE"), os.Getenv("BRANCH_NAME"))
			if err != nil {
				logger.Fatalf("Error committing and pushing changes: %v", err)
			}
		} else {
			logger.Println("📤 No changes to commit, skipping...")
		}
	} else {
		logger.Println("Skipping GitHub command functions in DRY_RUN mode")
	}

	logger.Printf("🚩 Execution Duration: %s\n", time.Since(start))
}

// setClock sets the clock and timezone
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

// withClock adds the clock to the context
func withClock(ctx context.Context, cl clock.Clock) context.Context {
	return context.WithValue(ctx, clock.ClockKey{}, cl)
}

// updateReadme updates the README.md file with the provided stats
func updateReadme(u, n string) error {
	f := "README.md"
	b, err := os.ReadFile("README.md")
	if err != nil {
		return err
	}

	if n == "" {
		n = "readme-stats"
	}

	s := fmt.Sprintf("<!--START_SECTION:%s-->", n)
	e := fmt.Sprintf("<!--END_SECTION:%s-->", n)

	si := strings.Index(string(b), s)
	ei := strings.Index(string(b), e)

	if si == -1 || ei == -1 {
		return fmt.Errorf("section tags %s or %s not found in %s", s, e, f)
	}

	u = string(b)[:si+len(s)] + "\n" + u + "\n" + string(b)[ei:]

	return os.WriteFile(f, []byte(u), 0644)
}
