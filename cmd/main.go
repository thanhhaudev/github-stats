package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	_ "time/tzdata"

	_ "github.com/joho/godotenv/autoload"
	"github.com/thanhhaudev/github-stats/pkg/clock"
	"github.com/thanhhaudev/github-stats/pkg/config"
	"github.com/thanhhaudev/github-stats/pkg/container"
	"github.com/thanhhaudev/github-stats/pkg/github"
	"github.com/thanhhaudev/github-stats/pkg/wakatime"
)

func main() {
	logger := log.New(os.Stdout, "", log.Lmsgprefix)
	logger.Println("🚀 Starting...")

	// load configuration
	cfg := config.Load()

	// validate configuration
	if err := cfg.Validate(); err != nil {
		logger.Fatalf("❌ Configuration error: %v", err)
	}

	cl, err := setClock(logger, cfg)
	if err != nil {
		panic(err)
	}

	start := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	ctx = withClock(ctx, cl)
	defer cancel()

	gc := github.NewGitHub(cfg.GitHubToken, cfg.Debug, cfg.HideRepoInfo)
	wc := wakatime.NewWakaTime(logger, cfg.WakaTimeAPIKey, wakatime.StatsRange(cfg.WakaTimeRange))
	dc := container.NewDataContainer(logger, container.NewClientManager(wc, gc), cfg)
	dc.SetClock(cl)
	if err := runGroupedStep(logger, "Build data container", cfg.EnableGitHubGroups, func() error {
		return dc.Build(ctx)
	}); err != nil {
		logger.Fatalln(err)
	}

	err = runGroupedStep(logger, "Update README", cfg.EnableGitHubGroups, func() error {
		logger.Println("📝 Updating README.md...")
		return updateReadme(dc.GetStats(cl), cfg.SectionName)
	})
	if err != nil {
		logger.Fatalf("Error updating README.md: %v", err)
	}

	if !cfg.DryRun {
		err = runGroupedStep(logger, "Configure git", cfg.EnableGitHubGroups, func() error {
			logger.Println("🔧 Setting up git config...")
			return setupGitConfig(
				dc.Data.Viewer.Login,
				cfg.GitHubToken,
				cfg.CommitUserName,
				cfg.CommitUserEmail,
				cfg.HideRepoInfo,
			)
		})
		if err != nil {
			logger.Fatalf("Error setting up git config: %v", err)
		}

		if cfg.EnableCache {
			if err := verifyCacheNotPushable(cfg.CacheFile); err != nil {
				logger.Fatalf("❌ Cache safety check failed: %v", err)
			}
		}

		err = runGroupedStep(logger, "Commit and push README", cfg.EnableGitHubGroups, func() error {
			changed, err := hasReadmeChanged()
			if err != nil {
				return err
			}

			if changed {
				logger.Println("📤 Committing and pushing changes...")
				return commitAndPushReadme(cfg.CommitMessage, cfg.BranchName, cfg.HideRepoInfo)
			}

			logger.Println("📤 No changes to commit, skipping...")
			return nil
		})
		if err != nil {
			logger.Fatalf("Error committing and pushing changes: %v", err)
		}
	} else {
		logger.Println("Skipping GitHub command functions in DRY_RUN mode")
	}

	logger.Printf("🚩 Execution Duration: %s\n", time.Since(start))
}

// setClock sets the clock and timezone
func setClock(logger *log.Logger, cfg *config.Config) (clock.Clock, error) {
	cl := clock.NewClock()
	if cfg.TimeZone != "" {
		err := cl.SetLocation(cfg.TimeZone)
		if err != nil {
			logger.Printf("⚠️ Invalid timezone %s: %v\n", cfg.TimeZone, err)

			return nil, err
		}

		if !cfg.SimpleLogs {
			logger.Printf("🕙 Timezone set to %s\n", cfg.TimeZone)
		}
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
