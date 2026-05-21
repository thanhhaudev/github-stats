package cache

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/thanhhaudev/github-stats/pkg/github"
	"github.com/thanhhaudev/github-stats/pkg/wakatime"
)

func tempCachePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "cache.json")
}

func TestLoad_MissingFileReturnsEmpty(t *testing.T) {
	c := Load(tempCachePath(t), false)

	if c == nil {
		t.Fatal("Load returned nil")
	}
	if len(c.Repos) != 0 {
		t.Errorf("expected empty repos, got %d", len(c.Repos))
	}
	if c.Version != RepoSchemaVersion {
		t.Errorf("expected version %d, got %d", RepoSchemaVersion, c.Version)
	}
}

func TestLoad_VersionMismatchReturnsEmpty(t *testing.T) {
	path := tempCachePath(t)
	stale := map[string]any{
		"version":        RepoSchemaVersion - 1,
		"onlyMainBranch": false,
		"repos":          map[string]any{"u1": map[string]any{}},
	}
	data, _ := json.Marshal(stale)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	c := Load(path, false)
	if len(c.Repos) != 0 {
		t.Errorf("expected empty repos after version mismatch, got %d", len(c.Repos))
	}
}

func TestLoad_OnlyMainBranchMismatchReturnsEmpty(t *testing.T) {
	path := tempCachePath(t)
	pushed := time.Now()
	c := &Cache{
		Version:        RepoSchemaVersion,
		OnlyMainBranch: true,
		Repos: map[string]*RepoEntry{
			"u1": {PushedAt: pushed, Commits: []github.Commit{{OID: "abc"}}},
		},
	}
	if err := c.Save(path); err != nil {
		t.Fatal(err)
	}

	loaded := Load(path, false) // toggled flag
	if len(loaded.Repos) != 0 {
		t.Errorf("expected empty repos after onlyMainBranch mismatch, got %d", len(loaded.Repos))
	}
}

func TestLoad_CorruptJSONReturnsEmpty(t *testing.T) {
	path := tempCachePath(t)
	if err := os.WriteFile(path, []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}

	c := Load(path, false)
	if len(c.Repos) != 0 {
		t.Errorf("expected empty repos after corrupt JSON, got %d", len(c.Repos))
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := tempCachePath(t)
	pushed := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)

	original := &Cache{
		Version:        RepoSchemaVersion,
		OnlyMainBranch: false,
		Repos: map[string]*RepoEntry{
			"https://github.com/alice/demo": {
				PushedAt: pushed,
				Commits:  []github.Commit{{OID: "abc", Additions: 10, Deletions: 2, CommittedDate: pushed}},
			},
		},
	}

	if err := original.Save(path); err != nil {
		t.Fatal(err)
	}

	loaded := Load(path, false)

	entry, ok := loaded.Repos["https://github.com/alice/demo"]
	if !ok {
		t.Fatal("repo entry missing after round-trip")
	}
	if !entry.PushedAt.Equal(pushed) {
		t.Errorf("pushedAt mismatch: got %v want %v", entry.PushedAt, pushed)
	}
	if len(entry.Commits) != 1 || entry.Commits[0].OID != "abc" {
		t.Errorf("commits not round-tripped: %+v", entry.Commits)
	}
}

func TestSaveLoadRoundTrip_WakaTime(t *testing.T) {
	path := tempCachePath(t)
	stats := &wakatime.Stats{}
	stats.Data.Status = "ok"
	stats.Data.Range = "last_7_days"
	stats.Data.Languages = []wakatime.StatsItem{{Name: "Go", Text: "1 hr", Hours: 1, Percent: 100}}
	allTime := &wakatime.AllTimeSinceTodayStats{}
	allTime.Data.Text = "10 hrs"

	original := &Cache{
		Version:        RepoSchemaVersion,
		OnlyMainBranch: false,
		Repos:          make(map[string]*RepoEntry),
	}
	original.SetWakaTime("last_7_days", stats, allTime)

	if err := original.Save(path); err != nil {
		t.Fatal(err)
	}

	loaded := Load(path, false)
	gotStats, gotAllTime, ok := loaded.LookupWakaTime("last_7_days")
	if !ok {
		t.Fatal("expected cached WakaTime data after round-trip")
	}
	if gotStats.Data.Range != "last_7_days" || gotStats.Data.Languages[0].Name != "Go" {
		t.Errorf("stats not round-tripped: %+v", gotStats)
	}
	if gotAllTime.Data.Text != "10 hrs" {
		t.Errorf("all-time stats not round-tripped: %+v", gotAllTime)
	}
}

func TestLookupWakaTime_MissesWhenRangeChanged(t *testing.T) {
	c := &Cache{Repos: make(map[string]*RepoEntry)}
	stats := &wakatime.Stats{}
	allTime := &wakatime.AllTimeSinceTodayStats{}
	c.SetWakaTime("last_7_days", stats, allTime)

	if _, _, ok := c.LookupWakaTime("last_30_days"); ok {
		t.Fatal("expected miss when cached WakaTime range differs")
	}
}

func TestSaveLoadRoundTrip_DoesNotLeakRepoMetadata(t *testing.T) {
	// Guard against accidentally re-introducing fields like IsPrivate or
	// repo Name in the cache file. The schema must stay minimal so a leaked
	// cache file does not expose private repo metadata.
	path := tempCachePath(t)
	c := &Cache{
		Version: RepoSchemaVersion,
		Repos: map[string]*RepoEntry{
			"https://github.com/alice/secret": {
				PushedAt: time.Now(),
				Commits:  []github.Commit{{OID: "x"}},
			},
		},
	}
	if err := c.Save(path); err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	forbidden := []string{`"isPrivate"`, `"isFork"`, `"languages"`, `"primaryLanguage"`, `"owner"`, `"name"`, `"viewer"`}
	for _, f := range forbidden {
		if bytes.Contains(raw, []byte(f)) {
			t.Errorf("cache file contains forbidden field %s — schema is leaking metadata", f)
		}
	}
}

func TestLookup(t *testing.T) {
	pushed := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	c := &Cache{
		Repos: map[string]*RepoEntry{
			"u1": {PushedAt: pushed, Commits: []github.Commit{{OID: "x"}}},
		},
	}

	t.Run("hit when pushedAt unchanged", func(t *testing.T) {
		commits, ok := c.Lookup("u1", pushed)
		if !ok || len(commits) != 1 || commits[0].OID != "x" {
			t.Errorf("expected hit, got ok=%v commits=%+v", ok, commits)
		}
	})

	t.Run("hit when pushedAt older (no new pushes)", func(t *testing.T) {
		_, ok := c.Lookup("u1", pushed.Add(-1*time.Hour))
		if !ok {
			t.Error("expected hit when fresh pushedAt is older than cached")
		}
	})

	t.Run("miss when pushedAt advanced", func(t *testing.T) {
		_, ok := c.Lookup("u1", pushed.Add(1*time.Hour))
		if ok {
			t.Error("expected miss when pushedAt advanced")
		}
	})

	t.Run("miss when repo not cached", func(t *testing.T) {
		_, ok := c.Lookup("unknown", pushed)
		if ok {
			t.Error("expected miss for unknown repo")
		}
	})
}

func TestSet_OverwritesExisting(t *testing.T) {
	c := &Cache{Repos: make(map[string]*RepoEntry)}
	pushed1 := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	pushed2 := pushed1.Add(time.Hour)

	c.Set("u1", pushed1, []github.Commit{{OID: "old"}})
	c.Set("u1", pushed2, []github.Commit{{OID: "new"}})

	if c.Repos["u1"].Commits[0].OID != "new" {
		t.Errorf("expected overwrite, got %+v", c.Repos["u1"].Commits)
	}
	if !c.Repos["u1"].PushedAt.Equal(pushed2) {
		t.Errorf("expected pushedAt updated, got %v", c.Repos["u1"].PushedAt)
	}
}

func TestPrune_DropsMissingURLs(t *testing.T) {
	now := time.Now()
	c := &Cache{
		Repos: map[string]*RepoEntry{
			"keep": {PushedAt: now},
			"drop": {PushedAt: now},
		},
	}

	c.Prune([]string{"keep"})

	if _, ok := c.Repos["drop"]; ok {
		t.Error("drop entry should have been pruned")
	}
	if _, ok := c.Repos["keep"]; !ok {
		t.Error("keep entry was incorrectly pruned")
	}
}

func TestPrune_EmptyKeepDropsAll(t *testing.T) {
	c := &Cache{
		Repos: map[string]*RepoEntry{"u1": {}, "u2": {}},
	}
	c.Prune(nil)
	if len(c.Repos) != 0 {
		t.Errorf("expected all entries pruned, got %d", len(c.Repos))
	}
}

func TestSet_ConcurrentSafe(t *testing.T) {
	c := &Cache{Repos: make(map[string]*RepoEntry)}
	pushed := time.Now()

	done := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func(id int) {
			defer func() { done <- struct{}{} }()
			url := "u" + string(rune('a'+id%26))
			c.Set(url, pushed, []github.Commit{{OID: url}})
			c.Lookup(url, pushed)
		}(i)
	}
	for i := 0; i < 50; i++ {
		<-done
	}
	// no race detector violations is the test
}

func TestSave_RaceWithSet(t *testing.T) {
	// Save must hold the mutex during marshal — otherwise concurrent Set
	// will trigger "concurrent map write" panic from json.Marshal walking
	// the map.
	c := &Cache{Repos: make(map[string]*RepoEntry)}
	path := tempCachePath(t)
	pushed := time.Now()

	stop := make(chan struct{})
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
				c.Set("u1", pushed, []github.Commit{{OID: "x"}})
			}
		}
	}()

	for i := 0; i < 20; i++ {
		if err := c.Save(path); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	}
	close(stop)
}

func TestLoad_WakaTimeSurvivesRepoSchemaMismatch(t *testing.T) {
	path := tempCachePath(t)
	raw := map[string]any{
		"version":        RepoSchemaVersion - 1, // stale repo-commit schema
		"onlyMainBranch": false,
		"repos":          map[string]any{"u1": map[string]any{}},
		"wakaTime": map[string]any{
			"version": WakaTimeSchemaVersion,
			"range":   "last_7_days",
			"stats":   map[string]any{"data": map[string]any{"range": "last_7_days"}},
		},
	}
	data, _ := json.Marshal(raw)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	c := Load(path, false)
	if len(c.Repos) != 0 {
		t.Errorf("expected repos dropped on schema mismatch, got %d", len(c.Repos))
	}
	if _, _, ok := c.LookupWakaTime("last_7_days"); !ok {
		t.Error("expected WakaTime snapshot to survive a repo-schema mismatch")
	}
}

func TestLoad_WakaTimeSurvivesOnlyMainBranchMismatch(t *testing.T) {
	path := tempCachePath(t)
	stats := &wakatime.Stats{}
	stats.Data.Range = "last_7_days"

	original := &Cache{
		Version:        RepoSchemaVersion,
		OnlyMainBranch: true,
		Repos:          make(map[string]*RepoEntry),
	}
	original.SetWakaTime("last_7_days", stats, nil)
	if err := original.Save(path); err != nil {
		t.Fatal(err)
	}

	c := Load(path, false) // toggled onlyMainBranch
	if _, _, ok := c.LookupWakaTime("last_7_days"); !ok {
		t.Error("expected WakaTime snapshot to survive an onlyMainBranch toggle")
	}
}

func TestLoad_LegacyWakaTimeEntryGrandfathered(t *testing.T) {
	path := tempCachePath(t)
	raw := map[string]any{
		"version":        RepoSchemaVersion,
		"onlyMainBranch": false,
		"repos":          map[string]any{},
		"wakaTime": map[string]any{ // no "version" field — written before this change
			"range": "last_7_days",
			"stats": map[string]any{"data": map[string]any{"range": "last_7_days"}},
		},
	}
	data, _ := json.Marshal(raw)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	c := Load(path, false)
	if _, _, ok := c.LookupWakaTime("last_7_days"); !ok {
		t.Error("expected a legacy versionless WakaTime entry to be adopted")
	}

	// The grandfathered entry must keep surviving once its version has been
	// normalized and persisted by a Save.
	if err := c.Save(path); err != nil {
		t.Fatal(err)
	}
	reloaded := Load(path, false)
	if _, _, ok := reloaded.LookupWakaTime("last_7_days"); !ok {
		t.Error("expected legacy entry to survive a save/reload cycle after grandfathering")
	}
}

func TestLoad_WakaTimeDroppedOnSchemaMismatch(t *testing.T) {
	path := tempCachePath(t)
	raw := map[string]any{
		"version":        RepoSchemaVersion,
		"onlyMainBranch": false,
		"repos":          map[string]any{},
		"wakaTime": map[string]any{
			"version": WakaTimeSchemaVersion + 1, // future, incompatible schema
			"range":   "last_7_days",
			"stats":   map[string]any{"data": map[string]any{"range": "last_7_days"}},
		},
	}
	data, _ := json.Marshal(raw)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	c := Load(path, false)
	if _, _, ok := c.LookupWakaTime("last_7_days"); ok {
		t.Error("expected WakaTime snapshot dropped on a schema-version mismatch")
	}
}
