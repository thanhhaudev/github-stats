package container

import (
	"context"
	"errors"
	"io"
	"log"
	"testing"

	"github.com/thanhhaudev/github-stats/pkg/cache"
	"github.com/thanhhaudev/github-stats/pkg/config"
	"github.com/thanhhaudev/github-stats/pkg/wakatime"
)

func TestRestoreCachedWakaTimeStats(t *testing.T) {
	stats := &wakatime.Stats{}
	stats.Data.Range = "last_7_days"
	stats.Data.Languages = []wakatime.StatsItem{{Name: "Go"}}
	allTime := &wakatime.AllTimeSinceTodayStats{}
	allTime.Data.Text = "10 hrs"

	c := &cache.Cache{Repos: make(map[string]*cache.RepoEntry)}
	c.SetWakaTime("last_7_days", stats, allTime)

	d := &DataContainer{
		Logger: log.New(io.Discard, "", 0),
		Config: &config.Config{
			WakaTimeRange: "last_7_days",
			SimpleLogs:    true,
		},
		Cache: c,
	}

	if !d.restoreCachedWakaTimeStats() {
		t.Fatal("expected cached WakaTime data to be restored")
	}
	if d.Data.WakaTime == nil || d.Data.WakaTime.Data.Languages[0].Name != "Go" {
		t.Fatalf("stats not restored: %+v", d.Data.WakaTime)
	}
	if d.Data.WakaTimeAllTime == nil || d.Data.WakaTimeAllTime.Data.Text != "10 hrs" {
		t.Fatalf("all-time stats not restored: %+v", d.Data.WakaTimeAllTime)
	}
}

func TestCacheWakaTimeStatsOnlyWhenCacheEnabled(t *testing.T) {
	stats := &wakatime.Stats{}
	stats.Data.Range = "last_7_days"
	d := &DataContainer{
		Config: &config.Config{WakaTimeRange: "last_7_days"},
	}
	d.Data.WakaTime = stats

	d.cacheWakaTimeStats()

	if d.Cache != nil {
		t.Fatal("cacheWakaTimeStats should not create a cache when caching is disabled")
	}
}

func TestInitWakaStatsRestoresCacheWhenStatsPending(t *testing.T) {
	cachedStats := &wakatime.Stats{}
	cachedStats.Data.Range = "last_7_days"
	cachedStats.Data.Languages = []wakatime.StatsItem{{Name: "Go"}}
	cachedAllTime := &wakatime.AllTimeSinceTodayStats{}
	cachedAllTime.Data.Text = "10 hrs"

	c := &cache.Cache{Repos: make(map[string]*cache.RepoEntry)}
	c.SetWakaTime("last_7_days", cachedStats, cachedAllTime)

	pendingStats := &wakatime.Stats{}
	pendingStats.Data.Status = "pending_update"
	d := NewDataContainer(
		log.New(io.Discard, "", 0),
		&fakeDataClientManager{wakaStats: pendingStats},
		&config.Config{WakaTimeRange: "last_7_days", SimpleLogs: true},
	)
	d.Cache = c

	if err := d.InitWakaStats(context.Background()); err != nil {
		t.Fatalf("InitWakaStats returned error: %v", err)
	}
	if d.Data.WakaTime == nil || d.Data.WakaTime.Data.Languages[0].Name != "Go" {
		t.Fatalf("expected cached WakaTime stats, got %+v", d.Data.WakaTime)
	}
	if d.Data.WakaTimeAllTime == nil || d.Data.WakaTimeAllTime.Data.Text != "10 hrs" {
		t.Fatalf("expected cached all-time stats, got %+v", d.Data.WakaTimeAllTime)
	}
}

func TestInitWakaStatsKeepsFreshStatsWhenAllTimePending(t *testing.T) {
	// Cached snapshot from an earlier run.
	cachedStats := &wakatime.Stats{}
	cachedStats.Data.Range = "last_7_days"
	cachedStats.Data.Languages = []wakatime.StatsItem{{Name: "Rust"}}
	cachedAllTime := &wakatime.AllTimeSinceTodayStats{}
	cachedAllTime.Data.Text = "99 hrs"

	c := &cache.Cache{Repos: make(map[string]*cache.RepoEntry)}
	c.SetWakaTime("last_7_days", cachedStats, cachedAllTime)

	// This run: main stats fetch succeeds, all-time fetch is still processing.
	freshStats := &wakatime.Stats{}
	freshStats.Data.Status = "ok"
	freshStats.Data.Range = "last_7_days"
	freshStats.Data.Languages = []wakatime.StatsItem{{Name: "Go"}}

	d := NewDataContainer(
		log.New(io.Discard, "", 0),
		&fakeDataClientManager{
			wakaStats:  freshStats,
			allTimeErr: wakatime.ErrStatsNotReady,
		},
		&config.Config{WakaTimeRange: "last_7_days", SimpleLogs: true},
	)
	d.Cache = c

	if err := d.InitWakaStats(context.Background()); err != nil {
		t.Fatalf("InitWakaStats returned error: %v", err)
	}
	// The fresh main stats must be kept, not overwritten by the cached snapshot.
	if d.Data.WakaTime == nil || d.Data.WakaTime.Data.Languages[0].Name != "Go" {
		t.Fatalf("expected fresh WakaTime stats kept, got %+v", d.Data.WakaTime)
	}
	// All-time was unavailable, so the cached all-time snapshot is reused.
	if d.Data.WakaTimeAllTime == nil || d.Data.WakaTimeAllTime.Data.Text != "99 hrs" {
		t.Fatalf("expected cached all-time reused, got %+v", d.Data.WakaTimeAllTime)
	}
}

func TestInitWakaStatsKeepsFreshStatsWhenAllTimeErrors(t *testing.T) {
	// Cached snapshot from an earlier run.
	cachedStats := &wakatime.Stats{}
	cachedStats.Data.Range = "last_7_days"
	cachedStats.Data.Languages = []wakatime.StatsItem{{Name: "Rust"}}
	cachedAllTime := &wakatime.AllTimeSinceTodayStats{}
	cachedAllTime.Data.Text = "99 hrs"

	c := &cache.Cache{Repos: make(map[string]*cache.RepoEntry)}
	c.SetWakaTime("last_7_days", cachedStats, cachedAllTime)

	// This run: main stats fetch succeeds, all-time fetch fails with a
	// generic (non-ErrStatsNotReady) error, with SimpleLogs disabled so the
	// error-logging sub-branch is exercised.
	freshStats := &wakatime.Stats{}
	freshStats.Data.Status = "ok"
	freshStats.Data.Range = "last_7_days"
	freshStats.Data.Languages = []wakatime.StatsItem{{Name: "Go"}}

	d := NewDataContainer(
		log.New(io.Discard, "", 0),
		&fakeDataClientManager{
			wakaStats:  freshStats,
			allTimeErr: errors.New("network timeout"),
		},
		&config.Config{WakaTimeRange: "last_7_days", SimpleLogs: false},
	)
	d.Cache = c

	if err := d.InitWakaStats(context.Background()); err != nil {
		t.Fatalf("InitWakaStats returned error: %v", err)
	}
	// A generic all-time error must not propagate and must not discard the
	// fresh main stats.
	if d.Data.WakaTime == nil || d.Data.WakaTime.Data.Languages[0].Name != "Go" {
		t.Fatalf("expected fresh WakaTime stats kept, got %+v", d.Data.WakaTime)
	}
	// The cached all-time snapshot is reused after a generic all-time error.
	if d.Data.WakaTimeAllTime == nil || d.Data.WakaTimeAllTime.Data.Text != "99 hrs" {
		t.Fatalf("expected cached all-time reused, got %+v", d.Data.WakaTimeAllTime)
	}
}
