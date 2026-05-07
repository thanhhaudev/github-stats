package container

import (
	"testing"

	"github.com/thanhhaudev/github-stats/pkg/github"
)

func TestCacheLogMessages(t *testing.T) {
	t.Run("hidden cache enabled log omits file and count", func(t *testing.T) {
		got := cacheEnabledLogMessage(true, "/tmp/cache.json", 33)
		if got != "📦 Cache enabled" {
			t.Fatalf("unexpected message: %q", got)
		}
	})

	t.Run("visible cache enabled log includes file and count", func(t *testing.T) {
		got := cacheEnabledLogMessage(false, "/tmp/cache.json", 33)
		want := "📦 Cache enabled (file=/tmp/cache.json, entries=33)"
		if got != want {
			t.Fatalf("unexpected message: want %q got %q", want, got)
		}
	})

	t.Run("hidden cache saved log omits count", func(t *testing.T) {
		got := cacheSavedLogMessage(true, 33)
		if got != "📦 Cache saved" {
			t.Fatalf("unexpected message: %q", got)
		}
	})

	t.Run("visible cache saved log includes count", func(t *testing.T) {
		got := cacheSavedLogMessage(false, 33)
		want := "📦 Cache saved (33 repos)"
		if got != want {
			t.Fatalf("unexpected message: want %q got %q", want, got)
		}
	})
}

func TestViewerFetchedLogMessage(t *testing.T) {
	v := &github.Viewer{Login: "alice", ID: "123"}

	if got := viewerFetchedLogMessage(true, v); got != "Successfully fetched viewer" {
		t.Fatalf("unexpected hidden message: %q", got)
	}

	if got := viewerFetchedLogMessage(false, v); got != "Successfully fetched viewer: alice (ID: 123)" {
		t.Fatalf("unexpected visible message: %q", got)
	}
}

func TestCommitsFetchLogMessage(t *testing.T) {
	if got := fetchingCommitsLogMessage(true, true, 33); got != "🔍 Fetching commits from repositories..." {
		t.Fatalf("unexpected hidden message: %q", got)
	}

	if got := fetchingCommitsLogMessage(false, true, 1); got != "🔍 Fetching commits from 1 repository..." {
		t.Fatalf("unexpected visible singular message: %q", got)
	}

	if got := fetchingCommitsLogMessage(false, true, 33); got != "🔍 Fetching commits from 33 repositories..." {
		t.Fatalf("unexpected visible plural message: %q", got)
	}
}
