package container

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/thanhhaudev/github-stats/pkg/cache"
	"github.com/thanhhaudev/github-stats/pkg/clock"
	"github.com/thanhhaudev/github-stats/pkg/config"
	"github.com/thanhhaudev/github-stats/pkg/github"
	"github.com/thanhhaudev/github-stats/pkg/wakatime"
	"github.com/thanhhaudev/github-stats/pkg/writer"
)

const (
	repoPerQuery   = 25
	branchPerQuery = 30
	commitPerQuery = 100
)

type DataContainer struct {
	ClientManager *ClientManager
	Logger        *log.Logger
	Config        *config.Config
	Cache         *cache.Cache // nil when caching is disabled
	Data          struct {
		Viewer          *github.Viewer
		Repositories    []github.Repository
		Commits         []github.Commit
		WakaTime        *wakatime.Stats
		WakaTimeAllTime *wakatime.AllTimeSinceTodayStats
	}
}

// metrics returns the metrics map
func (d *DataContainer) metrics(com *CommitStats, lang *LanguageStats, ai *AIStats) map[string]string {
	version := d.Config.ProgressBarVersion
	aiBlock := ""
	if ai != nil && ai.HasData {
		aiBlock = writer.MakeAIStatsList(ai.AIAdditions, ai.HumanAdditions, ai.AIInputTokens, ai.AIOutputTokens, ai.AvgPromptLength, d.Config.WakaTimeRange)
	}
	return map[string]string{
		config.MetricLanguagePerRepo:   writer.MakeLanguagePerRepoList(d.Data.Repositories, version),
		config.MetricLanguagesAndTools: writer.MakeLanguageAndToolList(lang.Languages, lang.TotalSize),
		config.MetricCommitDaysOfWeek:  writer.MakeCommitDaysOfWeekList(com.DailyCommits, com.TotalCommits, version),
		config.MetricCommitTimesOfDay:  writer.MakeCommitTimesOfDayList(d.Data.Commits, d.Config.SimplifyCommitTimesTitle, version),
		config.MetricWakaTimeSpentTime: writer.MakeWakaActivityList(
			d.Data.WakaTime,
			d.Config.WakaTimeData,
			version,
		),
		config.MetricCodingStreak:    writer.MakeCodingStreakList(d.Data.WakaTimeAllTime, com.CurrentStreak, com.LongestStreak),
		config.MetricWakaTimeAIStats: aiBlock,
	}
}

// GetStats returns the statistics
func (d *DataContainer) GetStats(c clock.Clock) string {
	b := strings.Builder{}

	// show metrics based on the environment variable
	w := d.metrics(d.CalculateCommits(), d.CalculateLanguages(), d.CalculateAIStats())
	for _, k := range d.Config.ShowMetrics {
		v, ok := w[k]
		if !ok {
			continue
		}

		b.WriteString(v)
	}

	// Show last update time if enabled
	showLastUpdated(c, &b, d.Config)

	d.Logger.Println("Created statistics successfully")

	return b.String()
}

func showLastUpdated(cl clock.Clock, b *strings.Builder, cfg *config.Config) {
	if cfg.ShowLastUpdate {
		layout := cfg.TimeLayout
		if layout == "" {
			layout = clock.DateTimeFormatWithTimezone
		}

		b.WriteString(writer.MakeLastUpdatedOn(cl.Now().Format(layout)))
	}
}

// InitViewer initializes the viewer
func (d *DataContainer) InitViewer(ctx context.Context) error {
	d.Logger.Println("Fetching viewer information...")

	v, err := d.ClientManager.GetViewer(ctx)
	if err != nil {
		return err
	}

	if v == nil {
		return fmt.Errorf("❌ could not fetch viewer information, please check your GitHub token")
	}

	d.Data.Viewer = v
	if d.Config.Debug {
		d.Logger.Printf("Successfully fetched viewer: %s (ID: %s)\n", d.Data.Viewer.Login, d.Data.Viewer.ID)
	}

	return nil
}

// InitRepositories initializes the repositories
// owned and contributed to by the user
func (d *DataContainer) InitRepositories(ctx context.Context) error {
	d.Logger.Println("Fetching repositories...")
	seenRepos := make(map[string]bool)
	errChan := make(chan error, 2)
	repoChan := make(chan []github.Repository, 2)

	go func() {
		r, err := d.ClientManager.GetOwnedRepositories(ctx, d.Data.Viewer.Login, repoPerQuery)
		if err != nil {
			errChan <- err
			return
		}

		repoChan <- r
		errChan <- nil

		d.Logger.Println("Fetched owned repositories successfully")
	}()

	go func() {
		c, err := d.ClientManager.GetContributedToRepositories(ctx, d.Data.Viewer.Login, repoPerQuery)
		if err != nil {
			errChan <- err
			return
		}

		repoChan <- c
		errChan <- nil

		d.Logger.Println("Fetched contributed to repositories successfully")
	}()

	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}

	close(repoChan) // Close the channel to signal that all repositories have been fetched

	// Deduplicate repositories
	for repos := range repoChan {
		for _, repo := range repos {
			if d.Config.ExcludeForkRepos && repo.IsFork {
				continue
			}

			if !seenRepos[repo.Url] {
				seenRepos[repo.Url] = true
				d.Data.Repositories = append(d.Data.Repositories, repo)
			}
		}
	}

	return nil
}

// InitCommits initializes the branches of the repositories
func (d *DataContainer) InitCommits(ctx context.Context) error {
	d.Logger.Println("Fetching commits...")
	fetchAllBranches := !d.Config.OnlyMainBranch
	hiddenRepoInfo := d.Config.HideRepoInfo
	repoCount := len(d.Data.Repositories)
	errChan := make(chan error, repoCount)
	commitChan := make(chan []github.Commit, repoCount)
	seenOIDs := make(map[string]bool)
	var mu sync.Mutex

	mask := func(input string) string {
		length := len(input)
		if length <= 2 {
			return input // No masking for very short strings
		}

		num := length / 3
		prefixLength := (length - num) / 2
		suffixLength := length - prefixLength - num

		return input[:prefixLength] + strings.Repeat("*", num) + input[len(input)-suffixLength:]
	}

	if hiddenRepoInfo {
		if d.Config.Debug {
			d.Logger.Printf("🔍 Fetching commits from %d %s...", repoCount, func() string {
				if repoCount == 1 {
					return "repository"
				}
				return "repositories"
			}())
		} else {
			d.Logger.Println("🔍 Fetching commits from repositories...")
		}
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // Limit to 5 concurrent goroutines

	for i, repo := range d.Data.Repositories {
		wg.Add(1)
		go func(i int, repo github.Repository) {
			defer wg.Done()
			progress := fmt.Sprintf("[%d/%d]", i+1, repoCount)

			// Skip the network round-trip when this repo has not been pushed to since the cached snapshot
			if d.Cache != nil {
				if cached, ok := d.Cache.Lookup(repo.Url, repo.PushedAt); ok {
					if !hiddenRepoInfo {
						d.Logger.Printf("%s Reusing %d cached commits: %s\n", progress, len(cached), mask(repo.Name))
					}
					commitChan <- cached
					errChan <- nil
					return
				}
			}

			var fetched []github.Commit

			if fetchAllBranches {
				if !hiddenRepoInfo {
					d.Logger.Printf("%s Fetching commits from all branches of: %s\n", progress, mask(repo.Name))
				}

				branches, err := d.ClientManager.GetBranches(ctx, repo.Owner.Login, repo.Name, branchPerQuery)
				if err != nil {
					errChan <- err
					return
				}

				var branchWg sync.WaitGroup
				for _, branch := range branches {
					branchWg.Add(1)
					semaphore <- struct{}{} // Acquire a slot
					go func(branch github.Branch) {
						defer branchWg.Done()
						defer func() { <-semaphore }() // Release the slot
						commits, err := d.ClientManager.GetCommits(ctx, repo.Owner.Login, repo.Name, d.Data.Viewer.ID, fmt.Sprintf("refs/heads/%s", branch.Name), commitPerQuery)
						if err != nil {
							errChan <- err
							return
						}

						mu.Lock()
						fetched = append(fetched, commits...)
						if !hiddenRepoInfo && d.Config.Debug {
							log.Printf("%s Fetched %d commits from branch %s", progress, len(commits), mask(branch.Name))
						}
						mu.Unlock()
					}(branch)
				}

				branchWg.Wait()
			} else {
				if !hiddenRepoInfo {
					d.Logger.Printf("%s Fetching commits from default branch of: %s\n", progress, mask(repo.Name))
				}

				defaultBranch, err := d.ClientManager.GetDefaultBranch(ctx, repo.Owner.Login, repo.Name)
				if err != nil {
					errChan <- err
					return
				}

				commits, err := d.ClientManager.GetCommits(ctx, repo.Owner.Login, repo.Name, d.Data.Viewer.ID, fmt.Sprintf("refs/heads/%s", defaultBranch.Name), commitPerQuery)
				if err != nil {
					errChan <- err
					return
				}

				fetched = commits
			}

			if d.Cache != nil {
				// Store raw GraphQL UTC timestamps; ToClockTz is applied later in the
				// dedup loop so a TIME_ZONE change between runs is honored without
				// invalidating the cache.
				d.Cache.Set(repo.Url, repo.PushedAt, fetched)
			}
			commitChan <- fetched
			errChan <- nil
		}(i, repo)
	}

	go func() {
		wg.Wait()
		close(commitChan)
		close(errChan)
	}()

	for i := 0; i < repoCount; i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}

	// Deduplicate commits
	for commits := range commitChan {
		for _, commit := range commits {
			if !seenOIDs[commit.OID] {
				seenOIDs[commit.OID] = true
				commit.CommittedDate = ctx.Value(clock.ClockKey{}).(clock.Clock).ToClockTz(commit.CommittedDate)
				d.Data.Commits = append(d.Data.Commits, commit)
			}
		}
	}

	d.Logger.Println("Fetched commits successfully")
	return nil
}

// InitWakaStats initializes the WakaTime statistics
func (d *DataContainer) InitWakaStats(ctx context.Context) error {
	d.Logger.Println("Fetching WakaTime statistics...")

	v, err := d.ClientManager.GetWakaTimeStats(ctx)
	if err != nil {
		return err
	}

	// if the status is not "ok", print the message
	if v.Data.Status != "ok" {
		switch v.Data.Status {
		case "pending_update":
			d.Logger.Println("WakaTime is not ready yet")
		default:
			d.Logger.Println("An error occurred while fetching WakaTime data:", v.Data.Status)

			return nil // Skip if the status is unknown
		}
	}

	d.Data.WakaTime = v

	// fetch all-time stats for streak calculation
	d.Logger.Println("Fetching WakaTime all-time statistics...")
	allTimeStats, err := d.ClientManager.GetWakaTimeAllTimeSinceToday(ctx)
	if err != nil {
		d.Logger.Println("An error occurred while fetching WakaTime all-time data:", err)
	} else {
		d.Data.WakaTimeAllTime = allTimeStats
	}

	return nil
}

// Build builds the data container
func (d *DataContainer) Build(ctx context.Context) error {
	d.Logger.Println("Building data container...")

	if d.Config.EnableCache {
		d.Cache = cache.Load(d.Config.CacheFile, d.Config.OnlyMainBranch)
		d.Logger.Printf("📦 Cache enabled (file=%s, entries=%d)", d.Config.CacheFile, len(d.Cache.Repos))
		// Ensure the cache file exists from the start so actions/cache@v4's post-step
		// can save it even if the action exits before the defer below runs.
		if err := d.Cache.Save(d.Config.CacheFile); err != nil {
			d.Logger.Printf("⚠️ Failed to initialize cache file: %v", err)
		}
	}

	// Save cache via defer so a transient error in InitCommits or InitWakaStats
	// does not throw away repo entries that were successfully refreshed earlier
	// in the run. We only save once we have a populated repo list to prune
	// against — saving with d.Data.Repositories empty would prune everything.
	defer func() {
		if d.Cache == nil || len(d.Data.Repositories) == 0 {
			return
		}
		urls := make([]string, 0, len(d.Data.Repositories))
		for _, r := range d.Data.Repositories {
			urls = append(urls, r.Url)
		}
		d.Cache.Prune(urls)
		if err := d.Cache.Save(d.Config.CacheFile); err != nil {
			d.Logger.Printf("⚠️ Failed to save cache: %v", err)
		} else {
			d.Logger.Printf("📦 Cache saved (%d repos)", len(d.Cache.Repos))
		}
	}()

	// if the GitHub client is not nil, initialize the viewer, repositories, and commits
	if d.ClientManager.GitHubClient != nil {
		d.Logger.Println("Fetching data from GitHub APIs...")
		err := d.InitViewer(ctx)
		if err != nil {
			return err
		}

		err = d.InitRepositories(ctx)
		if err != nil {
			return err
		}

		err = d.InitCommits(ctx)
		if err != nil {
			return err
		}

		d.Logger.Println("Fetching data from GitHub APIs successfully")
	} else {
		d.Logger.Println("⚠️ GitHub client is nil, skipping GitHub data fetching")
	}

	// if the WakaTime client is not nil, fetch data from WakaTime APIs
	if d.ClientManager.WakaTimeClient != nil {
		d.Logger.Println("Fetching data from Wakatime APIs...")
		err := d.InitWakaStats(ctx)
		if err != nil {
			return err
		}

		d.Logger.Println("Fetching data from Wakatime APIs successfully")
	}

	d.Logger.Println("Built data container successfully")

	return nil
}

// NewDataContainer creates a new DataContainer
func NewDataContainer(l *log.Logger, cm *ClientManager, cfg *config.Config) *DataContainer {
	return &DataContainer{
		Logger:        l,
		ClientManager: cm,
		Config:        cfg,
	}
}
