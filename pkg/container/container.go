package container

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

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
	ClientManager dataClientManager
	Logger        *log.Logger
	Config        *config.Config
	Clock         clock.Clock
	Cache         *cache.Cache // nil when caching is disabled
	Data          struct {
		Viewer          *github.Viewer
		Repositories    []github.Repository
		Commits         []github.Commit
		WakaTime        *wakatime.Stats
		WakaTimeAllTime *wakatime.AllTimeSinceTodayStats
	}
}

type dataClientManager interface {
	HasGitHubClient() bool
	HasWakaTimeClient() bool
	GetViewer(ctx context.Context) (*github.Viewer, error)
	GetOwnedRepositories(ctx context.Context, username string, numRepos int) ([]github.Repository, error)
	GetContributedToRepositories(ctx context.Context, username string, numRepos int) ([]github.Repository, error)
	GetBranches(ctx context.Context, owner, name string, numBranches int) ([]github.Branch, error)
	GetCommits(ctx context.Context, owner, name, authorID, branch string, numCommits int) ([]github.Commit, error)
	GetDefaultBranch(ctx context.Context, owner, name string) (*github.Branch, error)
	GetWakaTimeStats(ctx context.Context) (*wakatime.Stats, error)
	GetWakaTimeAllTimeSinceToday(ctx context.Context) (*wakatime.AllTimeSinceTodayStats, error)
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

	if !d.Config.SimpleLogs {
		d.Logger.Println("Created statistics successfully")
	}

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
	if !d.Config.SimpleLogs {
		d.Logger.Println("Fetching viewer information...")
	}

	v, err := d.ClientManager.GetViewer(ctx)
	if err != nil {
		return err
	}

	if v == nil {
		return fmt.Errorf("❌ could not fetch viewer information, please check your GitHub token")
	}

	d.Data.Viewer = v
	if d.Config.Debug && !d.Config.SimpleLogs {
		d.Logger.Println(viewerFetchedLogMessage(d.Config.HideRepoInfo, d.Data.Viewer))
	} else if d.Config.HideRepoInfo && !d.Config.SimpleLogs {
		d.Logger.Println(viewerFetchedLogMessage(true, d.Data.Viewer))
	}

	return nil
}

// InitRepositories initializes the repositories
// owned and contributed to by the user
func (d *DataContainer) InitRepositories(ctx context.Context) error {
	if !d.Config.SimpleLogs {
		d.Logger.Println("Fetching repositories...")
	}
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

		if !d.Config.SimpleLogs {
			d.Logger.Println("Fetched owned repositories successfully")
		}
	}()

	go func() {
		c, err := d.ClientManager.GetContributedToRepositories(ctx, d.Data.Viewer.Login, repoPerQuery)
		if err != nil {
			errChan <- err
			return
		}

		repoChan <- c
		errChan <- nil

		if !d.Config.SimpleLogs {
			d.Logger.Println("Fetched contributed to repositories successfully")
		}
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
	if !d.Config.SimpleLogs {
		d.Logger.Println("Fetching commits...")
	}
	fetchAllBranches := !d.Config.OnlyMainBranch
	hiddenRepoInfo := d.Config.HideRepoInfo
	repoCount := len(d.Data.Repositories)
	type commitResult struct {
		commits []github.Commit
		err     error
	}
	resultChan := make(chan commitResult, repoCount)
	seenOIDs := make(map[string]bool)

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
		if !d.Config.SimpleLogs {
			d.Logger.Println(fetchingCommitsLogMessage(true, d.Config.Debug, repoCount))
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
					if !hiddenRepoInfo && !d.Config.SimpleLogs {
						d.Logger.Printf("%s Reusing %d cached commits: %s\n", progress, len(cached), mask(repo.Name))
					}
					resultChan <- commitResult{commits: cached}
					return
				}
			}

			fetched, err := d.fetchRepoCommits(ctx, repo, progress, fetchAllBranches, hiddenRepoInfo, mask, semaphore)
			if err != nil {
				resultChan <- commitResult{err: err}
				return
			}

			if d.Cache != nil {
				// Store raw GraphQL UTC timestamps; ToClockTz is applied later in the
				// dedup loop so a TIME_ZONE change between runs is honored without
				// invalidating the cache.
				d.Cache.Set(repo.Url, repo.PushedAt, fetched)
			}
			resultChan <- commitResult{commits: fetched}
		}(i, repo)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		if result.err != nil {
			return result.err
		}
		for _, commit := range result.commits {
			if !seenOIDs[commit.OID] {
				seenOIDs[commit.OID] = true
				commit.CommittedDate = d.Clock.ToClockTz(commit.CommittedDate)
				d.Data.Commits = append(d.Data.Commits, commit)
			}
		}
	}

	if !d.Config.SimpleLogs {
		d.Logger.Println("Fetched commits successfully")
	}
	return nil
}

func (d *DataContainer) fetchRepoCommits(
	ctx context.Context,
	repo github.Repository,
	progress string,
	fetchAllBranches bool,
	hiddenRepoInfo bool,
	mask func(string) string,
	semaphore chan struct{},
) ([]github.Commit, error) {
	if fetchAllBranches {
		if !hiddenRepoInfo && !d.Config.SimpleLogs {
			d.Logger.Printf("%s Fetching commits from all branches of: %s\n", progress, mask(repo.Name))
		}

		branches, err := d.ClientManager.GetBranches(ctx, repo.Owner.Login, repo.Name, branchPerQuery)
		if err != nil {
			return nil, fmt.Errorf("fetch branches for repo %s: %w", repo.Name, err)
		}

		var (
			fetched []github.Commit
			mu      sync.Mutex
		)
		g, groupCtx := errgroup.WithContext(ctx)
		for _, branch := range branches {
			branch := branch
			g.Go(func() error {
				select {
				case semaphore <- struct{}{}:
					defer func() { <-semaphore }()
				case <-groupCtx.Done():
					return groupCtx.Err()
				}

				commits, err := d.ClientManager.GetCommits(groupCtx, repo.Owner.Login, repo.Name, d.Data.Viewer.ID, fmt.Sprintf("refs/heads/%s", branch.Name), commitPerQuery)
				if err != nil {
					return fmt.Errorf("fetch commits for repo %s branch %s: %w", repo.Name, branch.Name, err)
				}

				mu.Lock()
				fetched = append(fetched, commits...)
				if !hiddenRepoInfo && d.Config.Debug && !d.Config.SimpleLogs {
					log.Printf("%s Fetched %d commits from branch %s", progress, len(commits), mask(branch.Name))
				}
				mu.Unlock()
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return nil, err
		}

		return fetched, nil
	}

	if !hiddenRepoInfo && !d.Config.SimpleLogs {
		d.Logger.Printf("%s Fetching commits from default branch of: %s\n", progress, mask(repo.Name))
	}

	defaultBranch, err := d.ClientManager.GetDefaultBranch(ctx, repo.Owner.Login, repo.Name)
	if err != nil {
		return nil, fmt.Errorf("fetch default branch for repo %s: %w", repo.Name, err)
	}

	commits, err := d.ClientManager.GetCommits(ctx, repo.Owner.Login, repo.Name, d.Data.Viewer.ID, fmt.Sprintf("refs/heads/%s", defaultBranch.Name), commitPerQuery)
	if err != nil {
		return nil, fmt.Errorf("fetch commits for repo %s branch %s: %w", repo.Name, defaultBranch.Name, err)
	}

	return commits, nil
}

// InitWakaStats initializes the WakaTime statistics
func (d *DataContainer) InitWakaStats(ctx context.Context) error {
	if !d.Config.SimpleLogs {
		d.Logger.Println("Fetching WakaTime statistics...")
	}

	v, err := d.ClientManager.GetWakaTimeStats(ctx)
	if err != nil {
		if errors.Is(err, wakatime.ErrStatsNotReady) {
			d.restoreCachedWakaTimeStats()
			return nil
		}

		return err
	}

	// if the status is not "ok", print the message
	if v.Data.Status != "ok" {
		switch v.Data.Status {
		case "pending_update":
			if !d.Config.SimpleLogs {
				d.Logger.Println("WakaTime is not ready yet")
			}
			d.restoreCachedWakaTimeStats()
			return nil
		default:
			if !d.Config.SimpleLogs {
				d.Logger.Println("An error occurred while fetching WakaTime data:", v.Data.Status)
			}

			return nil // Skip if the status is unknown
		}
	}

	d.Data.WakaTime = v

	// fetch all-time stats for streak calculation
	if !d.Config.SimpleLogs {
		d.Logger.Println("Fetching WakaTime all-time statistics...")
	}
	allTimeStats, err := d.ClientManager.GetWakaTimeAllTimeSinceToday(ctx)
	if err != nil {
		if errors.Is(err, wakatime.ErrStatsNotReady) {
			d.restoreCachedWakaTimeStats()
			return nil
		}

		if !d.Config.SimpleLogs {
			d.Logger.Println("An error occurred while fetching WakaTime all-time data:", err)
		}
	} else {
		d.Data.WakaTimeAllTime = allTimeStats
	}

	d.cacheWakaTimeStats()

	return nil
}

func (d *DataContainer) restoreCachedWakaTimeStats() bool {
	if d.Cache == nil {
		return false
	}

	stats, allTime, ok := d.Cache.LookupWakaTime(d.Config.WakaTimeRange)
	if !ok {
		if !d.Config.SimpleLogs {
			d.Logger.Println("No cached WakaTime data available; skipping WakaTime metrics for this run")
		}
		return false
	}

	d.Data.WakaTime = stats
	d.Data.WakaTimeAllTime = allTime
	if !d.Config.SimpleLogs {
		d.Logger.Println("Reusing cached WakaTime data")
	}

	return true
}

func (d *DataContainer) cacheWakaTimeStats() {
	if d.Cache == nil || d.Data.WakaTime == nil {
		return
	}

	d.Cache.SetWakaTime(d.Config.WakaTimeRange, d.Data.WakaTime, d.Data.WakaTimeAllTime)
}

// Build builds the data container
func (d *DataContainer) Build(ctx context.Context) error {
	d.Logger.Println("Building data container...")

	if d.Config.EnableCache {
		d.Cache = cache.Load(d.Config.CacheFile, d.Config.OnlyMainBranch)
		if !d.Config.SimpleLogs {
			d.Logger.Println(cacheEnabledLogMessage(d.Config.HideRepoInfo, d.Config.CacheFile, len(d.Cache.Repos)))
		}
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
			if !d.Config.SimpleLogs {
				d.Logger.Println(cacheSavedLogMessage(d.Config.HideRepoInfo, len(d.Cache.Repos)))
			}
		}
	}()

	// if the GitHub client is not nil, initialize the viewer, repositories, and commits
	if d.ClientManager.HasGitHubClient() {
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

		if !d.Config.SimpleLogs {
			d.Logger.Println("Fetching data from GitHub APIs successfully")
		}
	} else {
		d.Logger.Println("⚠️ GitHub client is nil, skipping GitHub data fetching")
	}

	// if the WakaTime client is not nil, fetch data from WakaTime APIs
	if d.ClientManager.HasWakaTimeClient() {
		d.Logger.Println("Fetching data from Wakatime APIs...")
		err := d.InitWakaStats(ctx)
		if err != nil {
			return err
		}

		if !d.Config.SimpleLogs {
			d.Logger.Println("Fetching data from Wakatime APIs successfully")
		}
	}

	d.Logger.Println("Built data container successfully")

	return nil
}

// NewDataContainer creates a new DataContainer
func NewDataContainer(l *log.Logger, cm dataClientManager, cfg *config.Config) *DataContainer {
	return &DataContainer{
		Logger:        l,
		ClientManager: cm,
		Config:        cfg,
		Clock:         clock.NewClock(),
	}
}

func (d *DataContainer) SetClock(cl clock.Clock) {
	if cl == nil {
		return
	}

	d.Clock = cl
}

func cacheEnabledLogMessage(hidden bool, cachePath string, entryCount int) string {
	if hidden {
		return "📦 Cache enabled"
	}

	return fmt.Sprintf("📦 Cache enabled (file=%s, entries=%d)", cachePath, entryCount)
}

func cacheSavedLogMessage(hidden bool, count int) string {
	if hidden {
		return "📦 Cache saved"
	}

	if count == 1 {
		return "📦 Cache saved (1 repo)"
	}

	return fmt.Sprintf("📦 Cache saved (%d repos)", count)
}

func viewerFetchedLogMessage(hidden bool, v *github.Viewer) string {
	if hidden || v == nil {
		return "Successfully fetched viewer"
	}

	return fmt.Sprintf("Successfully fetched viewer: %s (ID: %s)", v.Login, v.ID)
}

func fetchingCommitsLogMessage(hidden bool, debug bool, repoCount int) string {
	if hidden {
		return "🔍 Fetching commits from repositories..."
	}

	if debug {
		if repoCount == 1 {
			return "🔍 Fetching commits from 1 repository..."
		}

		return fmt.Sprintf("🔍 Fetching commits from %d repositories...", repoCount)
	}

	return "🔍 Fetching commits from repositories..."
}
