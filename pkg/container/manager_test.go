package container

import (
	"context"
	"testing"

	"github.com/thanhhaudev/github-stats/pkg/github"
)

type fakeRepositoryService struct {
	branchRequests []map[string]interface{}
}

func (f *fakeRepositoryService) Branches(ctx context.Context, request *github.Request) (*github.Branches, error) {
	vars := make(map[string]interface{}, len(request.Vars()))
	for k, v := range request.Vars() {
		vars[k] = v
	}
	f.branchRequests = append(f.branchRequests, vars)

	if len(f.branchRequests) == 1 {
		return &github.Branches{
			Nodes:    []github.Branch{{Name: "main"}},
			PageInfo: github.PageInfo{EndCursor: "cursor-1", HasNextPage: true},
		}, nil
	}

	return &github.Branches{
		Nodes:    []github.Branch{{Name: "release"}},
		PageInfo: github.PageInfo{HasNextPage: false},
	}, nil
}

func (f *fakeRepositoryService) Commits(ctx context.Context, request *github.Request) (*github.Commits, error) {
	return nil, nil
}

func (f *fakeRepositoryService) Owned(ctx context.Context, request *github.Request) (*github.Repositories, error) {
	return nil, nil
}

func (f *fakeRepositoryService) ContributedTo(ctx context.Context, request *github.Request) (*github.Repositories, error) {
	return nil, nil
}

func (f *fakeRepositoryService) DefaultBranch(ctx context.Context, request *github.Request) (*github.Branch, error) {
	return nil, nil
}

func TestClientManagerGetBranchesPaginatesWithCursor(t *testing.T) {
	repos := &fakeRepositoryService{}
	cm := &ClientManager{repositories: repos}

	branches, err := cm.GetBranches(context.Background(), "acme", "repo-one", 1)
	if err != nil {
		t.Fatalf("GetBranches returned error: %v", err)
	}

	if len(branches) != 2 {
		t.Fatalf("expected two branches across pages, got %d", len(branches))
	}
	if len(repos.branchRequests) != 2 {
		t.Fatalf("expected two paginated requests, got %d", len(repos.branchRequests))
	}
	if _, ok := repos.branchRequests[0]["afterCursor"]; ok {
		t.Fatalf("first request should not include afterCursor: %+v", repos.branchRequests[0])
	}
	if got := repos.branchRequests[1]["afterCursor"]; got != "cursor-1" {
		t.Fatalf("second request afterCursor = %v, want cursor-1", got)
	}
}
