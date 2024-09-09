package github

import (
	"context"
	"time"
)

type RepositoryService struct {
	Client *Client
}

type Repository struct {
	Name            string `json:"name"`
	Url             string `json:"url"`
	IsPrivate       bool   `json:"isPrivate"`
	IsFork          bool   `json:"isFork"`
	PrimaryLanguage *struct {
		Name string `json:"name"`
	} `json:"primaryLanguage"`
	Owner struct {
		Login string `json:"login"`
	} `json:"owner"`
	Languages struct {
		Edges []struct {
			Node Language `json:"node"`
			Size int      `json:"size"`
		} `json:"edges"`
	}
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
	Additions     int       `json:"additions"`
	Deletions     int       `json:"deletions"`
	CommittedDate time.Time `json:"committedDate"`
	OID           string    `json:"oid"`
}

type Commits struct {
	Nodes    []Commit `json:"nodes"`
	PageInfo PageInfo `json:"pageInfo"`
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
