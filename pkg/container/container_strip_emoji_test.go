package container

import (
	"io"
	"log"
	"testing"
	"time"

	"github.com/thanhhaudev/github-stats/pkg/clock"
	"github.com/thanhhaudev/github-stats/pkg/config"
	"github.com/thanhhaudev/github-stats/pkg/github"
	"github.com/thanhhaudev/github-stats/pkg/writer"
)

// TestGetStats_StripEmojiAll asserts that when STRIP_EMOJI=ALL is active,
// the assembled output of DataContainer.GetStats contains zero emoji runes.
func TestGetStats_StripEmojiAll(t *testing.T) {
	cfg := &config.Config{
		GitHubToken: "ghp_test",
		ShowMetrics: []string{config.MetricCodingStreak},
		StripEmoji:  config.StripEmojiAll,
		SimpleLogs:  true,
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("cfg.Validate() unexpectedly failed: %v", err)
	}

	// Apply STRIP_EMOJI=ALL and restore the writer's global strip mode when the
	// test exits, so this setting does not leak into other tests.
	writer.SetStripEmoji(cfg.StripEmoji)
	t.Cleanup(func() { writer.SetStripEmoji("") })

	cl := clock.NewClock()
	dc := &DataContainer{
		Logger: log.New(io.Discard, "", 0),
		Config: cfg,
		Clock:  cl,
	}

	// Seed commits on today and yesterday so CalculateCommits returns
	// currentStreak=2 and longestStreak=2, which is the minimum needed for
	// MakeCodingStreakList to produce a non-empty markdown block.
	now := cl.Now()
	dc.Data.Commits = []github.Commit{
		{CommittedDate: now},
		{CommittedDate: now.AddDate(0, 0, -1)},
	}
	// WakaTimeAllTime is intentionally left nil; MakeCodingStreakList still
	// produces the streak lines when currentStreak > 0 || longestStreak > 0.

	stats := dc.GetStats(cl)

	if stats == "" {
		t.Fatal("GetStats returned empty string; expected non-empty markdown block from MakeCodingStreakList")
	}

	// Scan every rune and fail if any emoji rune survives STRIP_EMOJI=ALL.
	// The ranges mirror writer.isEmojiRune exactly.
	for i, r := range stats {
		if isEmojiRuneInTest(r) {
			t.Errorf(
				"found emoji rune %U at byte offset %d in output under STRIP_EMOJI=ALL; output:\n%s",
				r, i, stats,
			)
		}
	}
}

// isEmojiRuneInTest mirrors the ranges in writer.isEmojiRune so this test
// does not depend on an unexported function in another package.
func isEmojiRuneInTest(r rune) bool {
	switch {
	case r >= 0x1F300 && r <= 0x1FAFF:
		return true
	case r >= 0x2300 && r <= 0x23FF:
		return true
	case r >= 0x2600 && r <= 0x26FF:
		return true
	case r >= 0x2700 && r <= 0x27BF:
		return true
	case r >= 0x2B00 && r <= 0x2BFF:
		return true
	case r >= 0xFE00 && r <= 0xFE0F:
		return true
	case r == 0x200D:
		return true
	}
	return false
}

// Compile-time check: time is used for the Commit seeds above, ensure the
// import is not accidentally dropped if the seeding changes.
var _ = time.Now
