package container

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/thanhhaudev/github-stats/pkg/config"
	"github.com/thanhhaudev/github-stats/pkg/github"
	"github.com/thanhhaudev/github-stats/pkg/wakatime"
)

type fakeDataClientManager struct {
	mu         sync.Mutex
	branches   []github.Branch
	commitErr  error
	commitRefs []string
	wakaStats  *wakatime.Stats
	allTime    *wakatime.AllTimeSinceTodayStats
	allTimeErr error
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
