package container

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/thanhhaudev/github-stats/pkg/cache"
	"github.com/thanhhaudev/github-stats/pkg/config"
	"github.com/thanhhaudev/github-stats/pkg/github"
	"github.com/thanhhaudev/github-stats/pkg/wakatime"
)

type fakeDataClientManager struct {
	mu             sync.Mutex
	branches       []github.Branch
	commitErr      error
	commitRefs     []string
	wakaStats      *wakatime.Stats
	allTime        *wakatime.AllTimeSinceTodayStats
	allTimeErr     error
	commitFiles    map[string][]github.CommitFile
	commitFilesErr error
	commitFilesFn  func(owner, name, sha string) ([]github.CommitFile, error)
}

func (f *fakeDataClientManager) HasGitHubClient() bool {
	return true
}

func (f *fakeDataClientManager) HasWakaTimeClient() bool {
	return f.wakaStats != nil || f.allTime != nil
}

func (f *fakeDataClientManager) GetBranches(ctx context.Context, owner, name string, numBranches int) ([]github.Branch, error) {
	return f.branches, nil
}

func (f *fakeDataClientManager) GetCommits(ctx context.Context, owner, name, authorID, branch string, numCommits int) ([]github.Commit, error) {
	f.mu.Lock()
	f.commitRefs = append(f.commitRefs, branch)
	f.mu.Unlock()
	if branch == "refs/heads/fail" {
		return nil, f.commitErr
	}
	return []github.Commit{{OID: branch, CommittedDate: time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC)}}, nil
}

func (f *fakeDataClientManager) GetDefaultBranch(ctx context.Context, owner, name string) (*github.Branch, error) {
	return &github.Branch{Name: "main"}, nil
}

func (f *fakeDataClientManager) GetViewer(ctx context.Context) (*github.Viewer, error) {
	return &github.Viewer{ID: "viewer-id", Login: "viewer"}, nil
}

func (f *fakeDataClientManager) GetOwnedRepositories(ctx context.Context, username string, numRepos int) ([]github.Repository, error) {
	return nil, nil
}

func (f *fakeDataClientManager) GetContributedToRepositories(ctx context.Context, username string, numRepos int) ([]github.Repository, error) {
	return nil, nil
}

func (f *fakeDataClientManager) GetWakaTimeStats(ctx context.Context) (*wakatime.Stats, error) {
	return f.wakaStats, nil
}

func (f *fakeDataClientManager) GetWakaTimeAllTimeSinceToday(ctx context.Context) (*wakatime.AllTimeSinceTodayStats, error) {
	return f.allTime, f.allTimeErr
}

func (f *fakeDataClientManager) GetCommitFiles(ctx context.Context, owner, name, sha string) ([]github.CommitFile, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.commitFilesFn != nil {
		return f.commitFilesFn(owner, name, sha)
	}
	if f.commitFilesErr != nil {
		return nil, f.commitFilesErr
	}
	return f.commitFiles[sha], nil
}

func TestDataContainerInitCommitsReturnsBranchError(t *testing.T) {
	branchErr := errors.New("branch fetch failed")
	cm := &fakeDataClientManager{
		branches: []github.Branch{
			{Name: "main"},
			{Name: "fail"},
		},
		commitErr: branchErr,
	}
	cfg := &config.Config{OnlyMainBranch: false, SimpleLogs: true}
	d := NewDataContainer(log.Default(), cm, cfg)
	d.Data.Viewer = &github.Viewer{ID: "viewer-id"}
	repo := github.Repository{Name: "repo-one", Url: "https://github.com/acme/repo-one"}
	repo.Owner.Login = "acme"
	d.Data.Repositories = []github.Repository{repo}

	err := d.InitCommits(context.Background())

	if err == nil {
		t.Fatal("expected branch fetch error, got nil")
	}
	if !errors.Is(err, branchErr) {
		t.Fatalf("expected wrapped branch error, got %v", err)
	}
	if !strings.Contains(err.Error(), "repo-one") || !strings.Contains(err.Error(), "fail") {
		t.Fatalf("expected error to include repo and branch context, got %v", err)
	}
}

func TestDataContainerInitCommitsDoesNotRequireContextClock(t *testing.T) {
	cm := &fakeDataClientManager{}
	cfg := &config.Config{OnlyMainBranch: true, SimpleLogs: true}
	d := NewDataContainer(log.Default(), cm, cfg)
	d.Data.Viewer = &github.Viewer{ID: "viewer-id"}
	repo := github.Repository{Name: "repo-one", Url: "https://github.com/acme/repo-one"}
	repo.Owner.Login = "acme"
	d.Data.Repositories = []github.Repository{repo}

	if err := d.InitCommits(context.Background()); err != nil {
		t.Fatalf("InitCommits returned error: %v", err)
	}

	if len(d.Data.Commits) != 1 {
		t.Fatalf("expected one commit, got %d", len(d.Data.Commits))
	}
	if d.Data.Commits[0].OID != "refs/heads/main" {
		t.Fatalf("expected default branch commit, got %q", d.Data.Commits[0].OID)
	}
}

func TestInitCommitsEnrichesLanguagesAndCachesByOID(t *testing.T) {
	cm := &fakeDataClientManager{
		commitFiles: map[string][]github.CommitFile{
			"refs/heads/main": {{Filename: "main.go", Additions: 12}},
		},
	}
	cfg := &config.Config{OnlyMainBranch: true, SimpleLogs: true}
	d := NewDataContainer(log.Default(), cm, cfg)
	d.Data.Viewer = &github.Viewer{ID: "viewer-id"}
	repo := github.Repository{Name: "repo-one", Url: "https://github.com/acme/repo-one"}
	repo.Owner.Login = "acme"
	d.Data.Repositories = []github.Repository{repo}

	if err := d.InitCommits(context.Background()); err != nil {
		t.Fatalf("InitCommits error: %v", err)
	}

	if len(d.Data.Commits) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(d.Data.Commits))
	}
	c := d.Data.Commits[0]
	if c.Languages["Go"] != 12 {
		t.Errorf("expected Go=12 languages, got %v", c.Languages)
	}
	if c.RepoURL != repo.Url {
		t.Errorf("expected RepoURL %q, got %q", repo.Url, c.RepoURL)
	}
}

func TestInitCommitsRateLimitDegradesGracefully(t *testing.T) {
	cm := &fakeDataClientManager{commitFilesErr: github.ErrRateLimited}
	cfg := &config.Config{OnlyMainBranch: true, SimpleLogs: true}
	d := NewDataContainer(log.Default(), cm, cfg)
	d.Data.Viewer = &github.Viewer{ID: "viewer-id"}
	repo := github.Repository{Name: "repo-one", Url: "https://github.com/acme/repo-one"}
	repo.Owner.Login = "acme"
	d.Data.Repositories = []github.Repository{repo}

	if err := d.InitCommits(context.Background()); err != nil {
		t.Fatalf("rate limit should not fail the run, got: %v", err)
	}
	if len(d.Data.Commits) != 1 {
		t.Fatalf("expected the commit to still be recorded, got %d", len(d.Data.Commits))
	}
	if len(d.Data.Commits[0].Languages) != 0 {
		t.Errorf("expected empty languages under rate limit, got %v", d.Data.Commits[0].Languages)
	}
}

// TestInitCommitsRecoversAfterRateLimitedRunWithUnchangedPush verifies that a
// commit cached with empty languages during a rate-limited run is re-enriched on
// a later run with the SAME pushedAt (repo-cache hit), instead of being stuck
// empty forever.
func TestInitCommitsRecoversAfterRateLimitedRunWithUnchangedPush(t *testing.T) {
	shared := &cache.Cache{
		Repos:           map[string]*cache.RepoEntry{},
		CommitLanguages: map[string]map[string]int{},
	}
	cfg := &config.Config{OnlyMainBranch: true, SimpleLogs: true}
	repo := github.Repository{Name: "repo-one", Url: "https://github.com/acme/repo-one"}
	repo.Owner.Login = "acme"

	// Run 1: REST is rate-limited, so the single commit is cached with no languages.
	cm1 := &fakeDataClientManager{commitFilesErr: github.ErrRateLimited}
	d1 := NewDataContainer(log.Default(), cm1, cfg)
	d1.Cache = shared
	d1.Data.Viewer = &github.Viewer{ID: "viewer-id"}
	d1.Data.Repositories = []github.Repository{repo}
	if err := d1.InitCommits(context.Background()); err != nil {
		t.Fatalf("run 1 InitCommits error: %v", err)
	}
	if len(d1.Data.Commits[0].Languages) != 0 {
		t.Fatalf("run 1: expected empty languages, got %v", d1.Data.Commits[0].Languages)
	}

	// Run 2: SAME cache, SAME pushedAt (repo-cache hit), but REST now succeeds.
	cm2 := &fakeDataClientManager{
		commitFiles: map[string][]github.CommitFile{
			"refs/heads/main": {{Filename: "main.go", Additions: 12}},
		},
	}
	d2 := NewDataContainer(log.Default(), cm2, cfg)
	d2.Cache = shared
	d2.Data.Viewer = &github.Viewer{ID: "viewer-id"}
	d2.Data.Repositories = []github.Repository{repo}
	if err := d2.InitCommits(context.Background()); err != nil {
		t.Fatalf("run 2 InitCommits error: %v", err)
	}
	if len(d2.Data.Commits) != 1 {
		t.Fatalf("run 2: expected 1 commit, got %d", len(d2.Data.Commits))
	}
	if d2.Data.Commits[0].Languages["Go"] != 12 {
		t.Errorf("run 2: expected recovered Go=12, got %v", d2.Data.Commits[0].Languages)
	}
}

// TestInitCommitsPerRepoLanguagesSurviveGlobalDedup verifies that a commit shared
// by two repositories (same OID — e.g. an upstream the user also owns) is counted
// once in the totals but attributed to BOTH repos for the per-repo metric, rather
// than being collapsed nondeterministically to a single repo by global OID dedup.
func TestInitCommitsPerRepoLanguagesSurviveGlobalDedup(t *testing.T) {
	cm := &fakeDataClientManager{
		commitFiles: map[string][]github.CommitFile{
			// The fake's GetCommits returns OID == branch ref, so both repos
			// produce the same commit OID "refs/heads/main".
			"refs/heads/main": {{Filename: "main.go", Additions: 10}},
		},
	}
	cfg := &config.Config{OnlyMainBranch: true, SimpleLogs: true}
	d := NewDataContainer(log.Default(), cm, cfg)
	d.Data.Viewer = &github.Viewer{ID: "viewer-id"}
	repoA := github.Repository{Name: "repo-a", Url: "https://github.com/me/repo-a"}
	repoA.Owner.Login = "me"
	repoB := github.Repository{Name: "repo-b", Url: "https://github.com/up/repo-b"}
	repoB.Owner.Login = "up"
	d.Data.Repositories = []github.Repository{repoA, repoB}

	if err := d.InitCommits(context.Background()); err != nil {
		t.Fatalf("InitCommits error: %v", err)
	}

	// Global dedup: the shared OID is stored once for commit-count/total metrics.
	if len(d.Data.Commits) != 1 {
		t.Fatalf("expected 1 globally-deduped commit, got %d", len(d.Data.Commits))
	}

	lang := d.CalculateLanguages()
	if lang.TotalSize != 10 {
		t.Errorf("TotalSize = %d, want 10 (shared commit counted once)", lang.TotalSize)
	}
	if lang.RepoCount["Go"] != 2 {
		t.Errorf("RepoCount[Go] = %d, want 2 (both repos attributed); full=%v", lang.RepoCount["Go"], lang.RepoCount)
	}
}

// TestInitCommitsReconcilesPartiallyFailedSharedOID covers the case where the
// same commit OID exists in two repos but its file analysis succeeds in one and
// fails (transient error) in the other. After reconciliation both the totals and
// every repo's per-repo profile must use the successful result, deterministically.
func TestInitCommitsReconcilesPartiallyFailedSharedOID(t *testing.T) {
	cm := &fakeDataClientManager{
		commitFilesFn: func(owner, name, sha string) ([]github.CommitFile, error) {
			if owner == "me" { // succeeds for repo-a
				return []github.CommitFile{{Filename: "main.go", Additions: 10}}, nil
			}
			return nil, errors.New("transient fetch failure") // fails for repo-b
		},
	}
	cfg := &config.Config{OnlyMainBranch: true, SimpleLogs: true}
	d := NewDataContainer(log.Default(), cm, cfg)
	d.Data.Viewer = &github.Viewer{ID: "viewer-id"}
	repoA := github.Repository{Name: "repo-a", Url: "https://github.com/me/repo-a"}
	repoA.Owner.Login = "me"
	repoB := github.Repository{Name: "repo-b", Url: "https://github.com/up/repo-b"}
	repoB.Owner.Login = "up"
	d.Data.Repositories = []github.Repository{repoA, repoB}

	if err := d.InitCommits(context.Background()); err != nil {
		t.Fatalf("InitCommits error: %v", err)
	}

	lang := d.CalculateLanguages()
	// Totals must use the successful analysis regardless of which repo's result
	// reached the channel first.
	if lang.TotalSize != 10 {
		t.Errorf("TotalSize = %d, want 10 (reconciled from the successful repo)", lang.TotalSize)
	}
	if got := d.Data.Commits[0].Languages["Go"]; got != 10 {
		t.Errorf("global commit Languages[Go] = %d, want 10 (reconciled)", got)
	}
	// Both repos attributed Go even though repo-b's own fetch failed.
	if lang.RepoCount["Go"] != 2 {
		t.Errorf("RepoCount[Go] = %d, want 2 (both repos via reconciliation); full=%v", lang.RepoCount["Go"], lang.RepoCount)
	}
}
