package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/thanhhaudev/github-stats/pkg/github"
)

func tempCachePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "cache.json")
}

func sampleRepo(url string, pushed time.Time) github.Repository {
	r := github.Repository{
		Name:     "demo",
		Url:      url,
		PushedAt: pushed,
	}
	r.Owner.Login = "alice"
	return r
}

func TestLoad_MissingFileReturnsEmpty(t *testing.T) {
	c := Load(tempCachePath(t), false)

	if c == nil {
		t.Fatal("Load returned nil")
	}
	if len(c.Repos) != 0 {
		t.Errorf("expected empty repos, got %d", len(c.Repos))
	}
	if c.Version != SchemaVersion {
		t.Errorf("expected version %d, got %d", SchemaVersion, c.Version)
	}
}

func TestLoad_VersionMismatchReturnsEmpty(t *testing.T) {
	path := tempCachePath(t)
	stale := map[string]any{
		"version":        SchemaVersion - 1,
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
	c := &Cache{
		Version:        SchemaVersion,
		OnlyMainBranch: true,
		Repos: map[string]*RepoEntry{
			"u1": {Repo: sampleRepo("u1", time.Now()), Commits: []github.Commit{{OID: "abc"}}},
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
		Version:        SchemaVersion,
		OnlyMainBranch: false,
		Viewer:         &github.Viewer{Login: "alice", ID: "id1"},
		Repos: map[string]*RepoEntry{
			"https://github.com/alice/demo": {
				Repo:    sampleRepo("https://github.com/alice/demo", pushed),
				Commits: []github.Commit{{OID: "abc", Additions: 10, Deletions: 2, CommittedDate: pushed}},
			},
		},
	}

	if err := original.Save(path); err != nil {
		t.Fatal(err)
	}

	loaded := Load(path, false)

	if loaded.Viewer == nil || loaded.Viewer.Login != "alice" {
		t.Errorf("viewer not round-tripped: %+v", loaded.Viewer)
	}
	entry, ok := loaded.Repos["https://github.com/alice/demo"]
	if !ok {
		t.Fatal("repo entry missing after round-trip")
	}
	if !entry.Repo.PushedAt.Equal(pushed) {
		t.Errorf("pushedAt mismatch: got %v want %v", entry.Repo.PushedAt, pushed)
	}
	if len(entry.Commits) != 1 || entry.Commits[0].OID != "abc" {
		t.Errorf("commits not round-tripped: %+v", entry.Commits)
	}
}

func TestLookup(t *testing.T) {
	pushed := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	c := &Cache{
		Repos: map[string]*RepoEntry{
			"u1": {Repo: sampleRepo("u1", pushed), Commits: []github.Commit{{OID: "x"}}},
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

	c.Set("u1", sampleRepo("u1", pushed1), []github.Commit{{OID: "old"}})
	c.Set("u1", sampleRepo("u1", pushed2), []github.Commit{{OID: "new"}})

	if c.Repos["u1"].Commits[0].OID != "new" {
		t.Errorf("expected overwrite, got %+v", c.Repos["u1"].Commits)
	}
	if !c.Repos["u1"].Repo.PushedAt.Equal(pushed2) {
		t.Errorf("expected pushedAt updated, got %v", c.Repos["u1"].Repo.PushedAt)
	}
}

func TestPrune_DropsMissingURLs(t *testing.T) {
	c := &Cache{
		Repos: map[string]*RepoEntry{
			"keep": {Repo: sampleRepo("keep", time.Now())},
			"drop": {Repo: sampleRepo("drop", time.Now())},
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
			c.Set(url, sampleRepo(url, pushed), []github.Commit{{OID: url}})
			c.Lookup(url, pushed)
		}(i)
	}
	for i := 0; i < 50; i++ {
		<-done
	}
	// no race detector violations is the test
}
