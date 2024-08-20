package github

type RepositoryService struct {
	Client *Client
}

type Repository struct {
	Name            string `json:"name"`
	IsPrivate       bool   `json:"isPrivate"`
	IsFork          bool   `json:"isFork"`
	PrimaryLanguage *struct {
		Name string `json:"name"`
	} `json:"primaryLanguage"`
}

type Repositories struct {
	Nodes    []Repository `json:"nodes"`
	PageInfo PageInfo     `json:"pageInfo"`
}

// ContributedTo returns the repositories contributed to by the user
func (r *RepositoryService) ContributedTo(request *Request) (*Repositories, error) {
	var resp struct {
		Data struct {
			User struct {
				Repositories *Repositories `json:"repositoriesContributedTo"`
			} `json:"user"`
		} `json:"data"`
	}

	if err := r.Client.Post(request, &resp); err != nil {
		return nil, err
	}

	return resp.Data.User.Repositories, nil
}

// Owned returns the repositories owned by the user
func (r *RepositoryService) Owned(request *Request) (*Repositories, error) {
	var resp struct {
		Data struct {
			User struct {
				Repositories *Repositories `json:"repositories"`
			} `json:"user"`
		} `json:"data"`
	}

	if err := r.Client.Post(request, &resp); err != nil {
		return nil, err
	}

	return resp.Data.User.Repositories, nil
}
