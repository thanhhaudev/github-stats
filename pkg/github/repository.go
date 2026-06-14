package github

import (
	"context"
	"fmt"
	"time"
)

type RepositoryService struct {
	Client *Client
}

type Repository struct {
	Name      string    `json:"name"`
	Url       string    `json:"url"`
	IsPrivate bool      `json:"isPrivate"`
	IsFork    bool      `json:"isFork"`
	PushedAt  time.Time `json:"pushedAt"`
	Owner     struct {
		Login string `json:"login"`
	} `json:"owner"`
}

type Repositories struct {
	Nodes    []Repository `json:"nodes"`
	PageInfo PageInfo     `json:"pageInfo"`
}

// ContributedTo returns the repositories contributed to by the user
func (r *RepositoryService) ContributedTo(ctx context.Context, request *Request) (*Repositories, error) {
	var resp struct {
		Data struct {
			User struct {
				Repositories *Repositories `json:"repositoriesContributedTo"`
			} `json:"user"`
		} `json:"data"`
	}

	if err := r.Client.PostWithContext(ctx, request, "/graphql", &resp); err != nil {
		return nil, err
	}

	return resp.Data.User.Repositories, nil
}

// Owned returns the repositories owned by the user
func (r *RepositoryService) Owned(ctx context.Context, request *Request) (*Repositories, error) {
	var resp struct {
		Data struct {
			User struct {
				Repositories *Repositories `json:"repositories"`
			} `json:"user"`
		} `json:"data"`
	}

	if err := r.Client.PostWithContext(ctx, request, "/graphql", &resp); err != nil {
		return nil, err
	}

	return resp.Data.User.Repositories, nil
}

type Commit struct {
	Additions     int            `json:"additions"`
	Deletions     int            `json:"deletions"`
	CommittedDate time.Time      `json:"committedDate"`
	OID           string         `json:"oid"`
	RepoURL       string         `json:"repoURL,omitempty"`   // set in code, not from GraphQL
	Languages     map[string]int `json:"languages,omitempty"` // language -> additions, filled by enrichment
}

type Commits struct {
	Nodes    []Commit `json:"nodes"`
	PageInfo PageInfo `json:"pageInfo"`
}

// CommitFile is one entry of a REST commit's files[] array.
type CommitFile struct {
	Filename  string `json:"filename"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

// commitFilesCap mirrors GitHub's hard limit of 3000 files returned for a single
// commit across all pages.
const commitFilesCap = 3000

// CommitFiles fetches the files changed by a single commit via the REST API.
// GitHub returns up to 300 files in the first response and exposes the rest
// through Link-header pagination, up to a hard cap of 3000 files. We follow the
// rel="next" links so commits with more than 300 changed files are not silently
// undercounted.
func (r *RepositoryService) CommitFiles(ctx context.Context, owner, name, sha string) ([]CommitFile, error) {
	url := r.Client.origin + fmt.Sprintf("/repos/%s/%s/commits/%s", owner, name, sha)

	var files []CommitFile
	for url != "" {
		var page struct {
			Files []CommitFile `json:"files"`
		}

		header, err := r.Client.getURL(ctx, url, &page)
		if err != nil {
			return nil, err
		}

		files = append(files, page.Files...)
		if len(files) >= commitFilesCap {
			break
		}

		url = nextLink(header.Get("Link"))
	}

	return files, nil
}

// Commits returns the commits of a repository
func (r *RepositoryService) Commits(ctx context.Context, request *Request) (*Commits, error) {
	var resp struct {
		Data struct {
			Repository struct {
				Ref struct {
					Target struct {
						Commits *Commits `json:"history"`
					} `json:"target"`
				} `json:"ref"`
			} `json:"repository"`
		} `json:"data"`
	}

	if err := r.Client.PostWithContext(ctx, request, "/graphql", &resp); err != nil {
		return nil, err
	}

	return resp.Data.Repository.Ref.Target.Commits, nil
}

type Branch struct {
	Name string `json:"name"`
}

type Branches struct {
	Nodes    []Branch `json:"nodes"`
	PageInfo PageInfo `json:"pageInfo"`
}

// Branches returns the branches of a repository
func (r *RepositoryService) Branches(ctx context.Context, request *Request) (*Branches, error) {
	var resp struct {
		Data struct {
			Repository struct {
				Refs *Branches `json:"refs"`
			} `json:"repository"`
		} `json:"data"`
	}

	if err := r.Client.PostWithContext(ctx, request, "/graphql", &resp); err != nil {
		return nil, err
	}

	return resp.Data.Repository.Refs, nil
}

// DefaultBranch returns the default branch of a repository
func (r *RepositoryService) DefaultBranch(ctx context.Context, request *Request) (*Branch, error) {
	var resp struct {
		Data struct {
			Repository struct {
				DefaultBranchRef Branch `json:"defaultBranchRef"`
			} `json:"repository"`
		} `json:"data"`
	}

	if err := r.Client.PostWithContext(ctx, request, "/graphql", &resp); err != nil {
		return nil, err
	}

	return &resp.Data.Repository.DefaultBranchRef, nil
}
