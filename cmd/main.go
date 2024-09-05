package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
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
		logger.Fatalln("❌ GITHUB_TOKEN is required")
	}

	gc := github.NewGitHub(token)
	wc := wakatime.NewWakaTime(os.Getenv("WAKATIME_API_KEY"), wakatime.StatsRange(os.Getenv("WAKATIME_RANGE")))
	dc := container.NewDataContainer(logger, container.NewClientManager(wc, gc))
	if err := dc.Build(ctx); err != nil {
		logger.Fatalln(err)
	}

	logger.Println("🔧 Setting up git config...")
	err = setupGitConfig(dc.Data.Viewer.Login)
	if err != nil {
		logger.Fatalf("Error setting up git config: %v", err)
	}

	logger.Println("📝 Updating README.md...")
	err = updateReadme(dc.GetStats(cl), os.Getenv("SECTION_NAME"))
	if err != nil {
		logger.Fatalf("Error updating README.md: %v", err)
	}

	changed, err := hasReadmeChanged()
	if err != nil {
		logger.Fatalf("Error checking if README.md has changed: %v", err)
	}

	if changed {
		logger.Println("📤 Committing and pushing changes...")
		err = commitAndPushReadme("📝 Update README.md", "main")
		if err != nil {
			logger.Fatalf("Error committing and pushing changes: %v", err)
		}
	} else {
		logger.Println("📤 No changes to commit")
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

// setupGitConfig sets up the git configuration
func setupGitConfig(owner string) error {
	cmd := exec.Command("git", "config", "--global", "--add", "safe.directory", "/github/workspace")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git config safe.directory error: %v, output: %s", err, string(output))
	}

	cmd = exec.Command("git", "config", "--global", "user.name", "GitHub Action")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git config user.name error: %v, output: %s", err, string(output))
	}

	cmd = exec.Command("git", "config", "--global", "user.email", "action@github.com")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git config user.email error: %v, output: %s", err, string(output))
	}

	cmd = exec.Command("git", "remote", "set-url", "origin", fmt.Sprintf("https://%s@github.com/%s/%s.git", os.Getenv("GITHUB_TOKEN"), owner, owner))
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git remote set-url error: %v, output: %s", err, string(output))
	}

	return nil
}

// withClock adds the clock to the context
func withClock(ctx context.Context, cl clock.Clock) context.Context {
	return context.WithValue(ctx, clock.ClockKey{}, cl)
}

// hasReadmeChanged checks if README.md has changed
func hasReadmeChanged() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain", "README.md")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(string(output)) != "", nil
}

// commitAndPushReadme Commit and push changes if README.md has changed
func commitAndPushReadme(commitMessage, branch string) error {
	// Add the file to the staging area
	cmd := exec.Command("git", "add", "README.md")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	// Commit the changes
	cmd = exec.Command("git", "commit", "-m", commitMessage)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	// Push the changes
	cmd = exec.Command("git", "push", "origin", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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
