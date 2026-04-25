// Package cache persists fetched GitHub data between Action runs so we can
// skip re-fetching commits for repos whose pushedAt has not advanced.
//
// The file is intended to be restored/saved by actions/cache@v4 in the user's
// workflow. We do not commit it anywhere; if the file is missing or its schema
// is older than the binary, we fall back to a full fetch.
package cache

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"sync"
	"time"

	"github.com/thanhhaudev/github-stats/pkg/github"
)

// SchemaVersion bumps whenever the on-disk format changes incompatibly.
// A mismatch causes Load to return a fresh empty cache.
const SchemaVersion = 1

type RepoEntry struct {
	Repo    github.Repository `json:"repo"`
	Commits []github.Commit   `json:"commits"`
}

type Cache struct {
	Version        int                   `json:"version"`
	CachedAt       time.Time             `json:"cachedAt"`
	OnlyMainBranch bool                  `json:"onlyMainBranch"`
	Viewer         *github.Viewer        `json:"viewer,omitempty"`
	Repos          map[string]*RepoEntry `json:"repos"`

	mu sync.Mutex
}

// Load reads cache from path. Returns an empty cache (not nil) when the file
// is missing, malformed, version-mismatched, or built under a different
// onlyMainBranch flag — in those cases callers proceed with a full fetch.
func Load(path string, onlyMainBranch bool) *Cache {
	empty := &Cache{
		Version:        SchemaVersion,
		OnlyMainBranch: onlyMainBranch,
		Repos:          make(map[string]*RepoEntry),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// missing file is the expected first-run case, not an error
		if errors.Is(err, fs.ErrNotExist) {
			return empty
		}
		return empty
	}

	var c Cache
	if err := json.Unmarshal(data, &c); err != nil {
		return empty
	}

	if c.Version != SchemaVersion || c.OnlyMainBranch != onlyMainBranch {
		return empty
	}

	if c.Repos == nil {
		c.Repos = make(map[string]*RepoEntry)
	}

	return &c
}

// Save writes the cache atomically (write to tmp + rename) to avoid leaving
// a partial file if the process is killed mid-write.
func (c *Cache) Save(path string) error {
	c.mu.Lock()
	c.CachedAt = time.Now().UTC()
	c.Version = SchemaVersion
	c.mu.Unlock()

	data, err := json.MarshalIndent(c, "", "  ")
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
func (c *Cache) Lookup(repoURL string, pushedAt time.Time) ([]github.Commit, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.Repos[repoURL]
	if !ok {
		return nil, false
	}

	// pushedAt advances on any push; equal-or-older means nothing new
	if pushedAt.After(entry.Repo.PushedAt) {
		return nil, false
	}

	return entry.Commits, true
}

// Set stores fresh commits for a repo, overwriting any existing entry.
// Safe to call concurrently from goroutines.
func (c *Cache) Set(repoURL string, repo github.Repository, commits []github.Commit) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Repos[repoURL] = &RepoEntry{
		Repo:    repo,
		Commits: commits,
	}
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
