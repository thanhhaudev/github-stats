package github

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetWithContextDecodesJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Errorf("missing auth header, got %q", got)
		}
		_, _ = w.Write([]byte(`{"files":[{"filename":"a.go","additions":3,"deletions":1}]}`))
	}))
	defer srv.Close()

	c := NewClient("tok", false, false)
	c.origin = srv.URL

	var detail struct {
		Files []CommitFile `json:"files"`
	}
	if err := c.GetWithContext(context.Background(), "/repos/o/n/commits/sha", &detail); err != nil {
		t.Fatalf("GetWithContext error: %v", err)
	}
	if len(detail.Files) != 1 || detail.Files[0].Filename != "a.go" || detail.Files[0].Additions != 3 {
		t.Fatalf("unexpected decode: %+v", detail.Files)
	}
}

func TestGetWithContextReportsRateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	c := NewClient("tok", false, false)
	c.origin = srv.URL

	var detail struct{}
	err := c.GetWithContext(context.Background(), "/repos/o/n/commits/sha", &detail)
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}
}

func TestGetWithContextDetectsSecondaryRateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Secondary rate limit: 403 with Retry-After but a non-zero remaining.
		w.Header().Set("Retry-After", "60")
		w.Header().Set("X-RateLimit-Remaining", "42")
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	c := NewClient("tok", false, false)
	c.origin = srv.URL

	var detail struct{}
	err := c.GetWithContext(context.Background(), "/repos/o/n/commits/sha", &detail)
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("expected ErrRateLimited for secondary limit, got %v", err)
	}
}

func TestCommitFilesFollowsLinkPagination(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") == "2" {
			_, _ = w.Write([]byte(`{"files":[{"filename":"b.py","additions":2}]}`))
			return
		}
		// First page advertises a next page via the Link header (RFC 5988),
		// exactly like GitHub does for commits with more than 300 files.
		w.Header().Set("Link", "<"+srv.URL+"/next?page=2>; rel=\"next\", <"+srv.URL+"/next?page=2>; rel=\"last\"")
		_, _ = w.Write([]byte(`{"files":[{"filename":"a.go","additions":1}]}`))
	}))
	defer srv.Close()

	c := NewClient("tok", false, false)
	c.origin = srv.URL
	svc := &RepositoryService{Client: c}

	files, err := svc.CommitFiles(context.Background(), "o", "n", "sha")
	if err != nil {
		t.Fatalf("CommitFiles error: %v", err)
	}
	if len(files) != 2 || files[0].Filename != "a.go" || files[1].Filename != "b.py" {
		t.Fatalf("expected files from both pages, got %+v", files)
	}
}
