package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/thanhhaudev/github-stats/pkg/wakatime"
)

// Valid metric keys for SHOW_METRICS
const (
	MetricLanguagePerRepo   = "LANGUAGE_PER_REPO"
	MetricLanguagesAndTools = "LANGUAGES_AND_TOOLS"
	MetricCommitDaysOfWeek  = "COMMIT_DAYS_OF_WEEK"
	MetricCommitTimesOfDay  = "COMMIT_TIMES_OF_DAY"
	MetricWakaTimeSpentTime = "WAKATIME_SPENT_TIME"
	MetricCodingStreak      = "CODING_STREAK"
)

// Valid data types for WAKATIME_DATA
const (
	WakaDataEditors          = "EDITORS"
	WakaDataLanguages        = "LANGUAGES"
	WakaDataProjects         = "PROJECTS"
	WakaDataOperatingSystems = "OPERATING_SYSTEMS"
)

// Valid progress bar versions
const (
	ProgressBarVersion1 = "1"
	ProgressBarVersion2 = "2"
)

// Boolean string values
const (
	TrueVal  = "true"
	FalseVal = "false"
)

// Config holds all environment variables used in the application
type Config struct {
	// GitHub settings
	GitHubToken string

	// WakaTime settings
	WakaTimeAPIKey string
	WakaTimeRange  string
	WakaTimeData   []string

	// Display settings
	ShowMetrics              []string
	ShowLastUpdate           bool
	TimeLayout               string
	TimeZone                 string
	ProgressBarVersion       string
	SimplifyCommitTimesTitle bool

	// Git settings
	DryRun          bool
	CommitUserName  string
	CommitUserEmail string
	CommitMessage   string
	BranchName      string
	SectionName     string

	// Repository settings
	HideRepoInfo     bool
	ExcludeForkRepos bool
	OnlyMainBranch   bool

	// Cache settings
	EnableCache bool
	CacheTTL    int // in hours
	CacheFile   string
}

// Load reads all environment variables and returns a Config struct
func Load() *Config {
	cfg := &Config{
		// GitHub settings
		GitHubToken: os.Getenv("GITHUB_TOKEN"),

		// WakaTime settings
		WakaTimeAPIKey: os.Getenv("WAKATIME_API_KEY"),
		WakaTimeRange:  os.Getenv("WAKATIME_RANGE"),
		WakaTimeData:   splitEnv("WAKATIME_DATA"),

		// Display settings
		ShowMetrics:              splitEnv("SHOW_METRICS"),
		ShowLastUpdate:           os.Getenv("SHOW_LAST_UPDATE") == TrueVal,
		TimeLayout:               os.Getenv("TIME_LAYOUT"),
		TimeZone:                 os.Getenv("TIME_ZONE"),
		ProgressBarVersion:       os.Getenv("PROGRESS_BAR_VERSION"),
		SimplifyCommitTimesTitle: os.Getenv("SIMPLIFY_COMMIT_TIMES_TITLE") == TrueVal,

		// Git settings
		DryRun:          os.Getenv("DRY_RUN") == TrueVal,
		CommitUserName:  os.Getenv("COMMIT_USER_NAME"),
		CommitUserEmail: os.Getenv("COMMIT_USER_EMAIL"),
		CommitMessage:   os.Getenv("COMMIT_MESSAGE"),
		BranchName:      os.Getenv("BRANCH_NAME"),
		SectionName:     os.Getenv("SECTION_NAME"),

		// Repository settings
		HideRepoInfo:     os.Getenv("HIDE_REPO_INFO") == TrueVal,
		ExcludeForkRepos: os.Getenv("EXCLUDE_FORK_REPOS") == TrueVal,
		OnlyMainBranch:   os.Getenv("ONLY_MAIN_BRANCH") == TrueVal,

		// Cache settings
		EnableCache: os.Getenv("ENABLE_CACHE") == TrueVal,
		CacheFile:   os.Getenv("CACHE_FILE"),
	}

	if ttl := os.Getenv("CACHE_TTL"); ttl != "" {
		fmt.Sscanf(ttl, "%d", &cfg.CacheTTL)
	}

	cfg.applyDefaults()

	return cfg
}

// applyDefaults sets default values for optional configuration fields
func (c *Config) applyDefaults() {
	if c.WakaTimeAPIKey != "" && c.WakaTimeRange == "" {
		c.WakaTimeRange = string(wakatime.StatsRangeLast7Days)
	}

	if c.BranchName == "" {
		c.BranchName = "main"
	}

	if c.CommitMessage == "" {
		c.CommitMessage = "📝 Update README.md"
	}

	if c.SectionName == "" {
		c.SectionName = "readme-stats"
	}

	if c.CacheFile == "" {
		c.CacheFile = ".github-stats-cache.json"
	}

	if c.CacheTTL == 0 {
		c.CacheTTL = 2 // Default to 2 hours
	}
}

// splitEnv splits a comma-separated environment variable into a slice
func splitEnv(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return []string{}
	}
	return strings.Split(value, ",")
}

// Validate validates the configuration and returns an error if any value is invalid
func (c *Config) Validate() error {
	if c.GitHubToken == "" {
		return fmt.Errorf("GITHUB_TOKEN is required")
	}

	if c.WakaTimeAPIKey != "" && c.WakaTimeRange != "" {
		if !wakatime.StatsRange(c.WakaTimeRange).IsValid() {
			validRanges := []string{
				string(wakatime.StatsRangeLast7Days),
				string(wakatime.StatsRangeLast30Days),
				string(wakatime.StatsRangeLast6Months),
				string(wakatime.StatsLastYear),
				string(wakatime.StatsRangeAllTime),
			}
			return fmt.Errorf("WAKATIME_RANGE must be one of: %s (got: %s)", strings.Join(validRanges, ", "), c.WakaTimeRange)
		}
	}

	if c.WakaTimeAPIKey != "" && len(c.WakaTimeData) > 0 {
		validData := []string{
			WakaDataEditors,
			WakaDataLanguages,
			WakaDataProjects,
			WakaDataOperatingSystems,
		}
		for _, data := range c.WakaTimeData {
			trimmed := strings.TrimSpace(data)
			if trimmed != "" && !contains(validData, trimmed) {
				return fmt.Errorf("WAKATIME_DATA contains invalid value '%s'. Valid values: %s", trimmed, strings.Join(validData, ", "))
			}
		}
	}

	if len(c.ShowMetrics) == 0 {
		return fmt.Errorf("SHOW_METRICS is required")
	}

	validMetrics := []string{
		MetricLanguagePerRepo,
		MetricLanguagesAndTools,
		MetricCommitDaysOfWeek,
		MetricCommitTimesOfDay,
		MetricWakaTimeSpentTime,
		MetricCodingStreak,
	}
	for _, metric := range c.ShowMetrics {
		trimmed := strings.TrimSpace(metric)
		if trimmed != "" && !contains(validMetrics, trimmed) {
			return fmt.Errorf("SHOW_METRICS contains invalid value '%s'. Valid values: %s", trimmed, strings.Join(validMetrics, ", "))
		}
	}

	if c.ProgressBarVersion != "" && c.ProgressBarVersion != ProgressBarVersion1 && c.ProgressBarVersion != ProgressBarVersion2 {
		return fmt.Errorf("PROGRESS_BAR_VERSION must be '%s' or '%s' (got: %s)", ProgressBarVersion1, ProgressBarVersion2, c.ProgressBarVersion)
	}

	return nil
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
