// Package cache persists fetched GitHub and WakaTime data between Action runs
// so we can skip re-fetching commits for repos whose pushedAt has not advanced
// and replay the last ready WakaTime stats when the API is still processing.
//
// The file is intended to be restored/saved by actions/cache@v4 in the user's
// workflow. We do not commit it anywhere; if the file is missing or its schema
// is older than the binary, we fall back to a full fetch.
package cache

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/thanhhaudev/github-stats/pkg/github"
	"github.com/thanhhaudev/github-stats/pkg/wakatime"
)

// RepoSchemaVersion bumps whenever the repo-commit on-disk format changes
// incompatibly. A mismatch drops the cached repos (one slow run after the
// upgrade) but leaves the WakaTime snapshot intact.
const RepoSchemaVersion = 2

// WakaTimeSchemaVersion bumps whenever the WakaTime snapshot on-disk format
// changes incompatibly. It is independent of RepoSchemaVersion so a repo-commit
// format change never discards a still-valid WakaTime snapshot.
//
// Load maps legacy versionless entries to version 1; raising this constant
// past 1 therefore correctly invalidates them.
const WakaTimeSchemaVersion = 1

// RepoEntry holds the minimum data needed to decide cache hit/miss and replay
// a fetch result. We deliberately do NOT store the full Repository struct —
// fields like IsPrivate, Languages, Owner, Name reduce blast radius if the
// cache file ever leaks (e.g. accidental commit by the user).
type RepoEntry struct {
	PushedAt time.Time       `json:"pushedAt"`
	Commits  []github.Commit `json:"commits"`
}

type WakaTimeEntry struct {
	Version  int                              `json:"version"`
	CachedAt time.Time                        `json:"cachedAt"`
	Range    string                           `json:"range"`
	Stats    *wakatime.Stats                  `json:"stats"`
	AllTime  *wakatime.AllTimeSinceTodayStats `json:"allTime"`
}

type Cache struct {
	Version        int                   `json:"version"`
	CachedAt       time.Time             `json:"cachedAt"`
	OnlyMainBranch bool                  `json:"onlyMainBranch"`
	Repos          map[string]*RepoEntry `json:"repos"`
	WakaTime       *WakaTimeEntry        `json:"wakaTime,omitempty"`

	mu sync.Mutex
}

// Load reads cache from path. The repo-commit section and the WakaTime section
// are validated independently: a repo-schema or onlyMainBranch mismatch drops
// only the cached repos, while a WakaTime-schema mismatch drops only the
// WakaTime snapshot. A missing or malformed file yields an empty cache.
func Load(path string, onlyMainBranch bool) *Cache {
	fresh := &Cache{
		Version:        RepoSchemaVersion,
		OnlyMainBranch: onlyMainBranch,
		Repos:          make(map[string]*RepoEntry),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// missing file is the expected first-run case, not an error
		return fresh
	}

	var c Cache
	if err := json.Unmarshal(data, &c); err != nil {
		return fresh
	}

	// Repo-commit section: usable only when its schema and branch mode match.
	if c.Version != RepoSchemaVersion || c.OnlyMainBranch != onlyMainBranch || c.Repos == nil {
		c.Repos = make(map[string]*RepoEntry)
	}

	// WakaTime section: validated on its own version, independent of the repo
	// schema and branch mode. Entries written before WakaTime carried a version
	// field share the v1 on-disk shape, so a missing version is treated as 1.
	if c.WakaTime != nil {
		entryVersion := c.WakaTime.Version
		// version == 0 means the entry predates WakaTimeSchemaVersion; those
		// entries share the v1 on-disk layout, so treat them as v1. Bumping
		// WakaTimeSchemaVersion past 1 therefore correctly invalidates them.
		if entryVersion == 0 {
			entryVersion = 1
		}
		if entryVersion == WakaTimeSchemaVersion {
			// Stamp the canonical version so the next Save persists it.
			c.WakaTime.Version = WakaTimeSchemaVersion
		} else {
			c.WakaTime = nil
		}
	}

	// Normalize identity fields so the next Save writes the current schema.
	c.Version = RepoSchemaVersion
	c.OnlyMainBranch = onlyMainBranch

	return &c
}

// Save writes the cache atomically (write to tmp + rename) to avoid leaving
// a partial file if the process is killed mid-write. The mutex is held during
// the marshal so concurrent Set/Prune from goroutines cannot race the encoder.
func (c *Cache) Save(path string) error {
	c.mu.Lock()
	c.CachedAt = time.Now().UTC()
	c.Version = RepoSchemaVersion
	data, err := json.MarshalIndent(c, "", "  ")
	c.mu.Unlock()
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}

// Lookup returns cached commits when the repo's pushedAt has not advanced
// past the cached value. A returned ok=false means the caller must fetch fresh.
//
// Note: pushedAt has 1-second granularity. A push that lands in the same
// second as the cache write may be missed until the next run — acceptable
// trade-off for a daily-cadence Action.
func (c *Cache) Lookup(repoURL string, pushedAt time.Time) ([]github.Commit, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.Repos[repoURL]
	if !ok {
		return nil, false
	}

	if pushedAt.After(entry.PushedAt) {
		return nil, false
	}

	return entry.Commits, true
}

// Set stores fresh commits for a repo, overwriting any existing entry.
// Safe to call concurrently from goroutines.
//
// Cached commits are stored with their raw GraphQL UTC timestamps; timezone
// conversion (ToClockTz) is re-applied downstream so a TIME_ZONE change
// between runs does not require cache invalidation.
func (c *Cache) Set(repoURL string, pushedAt time.Time, commits []github.Commit) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Repos[repoURL] = &RepoEntry{
		PushedAt: pushedAt,
		Commits:  commits,
	}
}

func (c *Cache) SetWakaTime(statsRange string, stats *wakatime.Stats, allTime *wakatime.AllTimeSinceTodayStats) {
	if stats == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.WakaTime = &WakaTimeEntry{
		Version:  WakaTimeSchemaVersion,
		CachedAt: time.Now().UTC(),
		Range:    statsRange,
		Stats:    stats,
		AllTime:  allTime,
	}
}

func (c *Cache) LookupWakaTime(statsRange string) (*wakatime.Stats, *wakatime.AllTimeSinceTodayStats, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.WakaTime == nil || c.WakaTime.Stats == nil || c.WakaTime.Range != statsRange {
		return nil, nil, false
	}

	return c.WakaTime.Stats, c.WakaTime.AllTime, true
}

// Prune removes entries whose URL is not in keepURLs. Used to drop cache for
// repos that have been deleted or transferred so they stop inflating stats.
func (c *Cache) Prune(keepURLs []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	keep := make(map[string]struct{}, len(keepURLs))
	for _, u := range keepURLs {
		keep[u] = struct{}{}
	}

	for url := range c.Repos {
		if _, ok := keep[url]; !ok {
			delete(c.Repos, url)
		}
	}
}
