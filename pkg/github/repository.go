package github

type RepositoryService struct {
	Client *Client
}

type Repository struct {
	Name            string `json:"name"`
	IsPrivate       bool   `json:"isPrivate"`
	IsFork          bool   `json:"isFork"`
	PrimaryLanguage struct {
		Name string `json:"name"`
	} `json:"primaryLanguage"`
}

// ContributedTo returns the repositories contributed to by the user
func (r *RepositoryService) ContributedTo(request *Request) ([]Repository, error) {
	var resp struct {
		Data struct {
			User struct {
				Repositories struct {
					Nodes []Repository `json:"nodes"`
				} `json:"repositoriesContributedTo"`
			} `json:"user"`
		} `json:"data"`
	}

	if err := r.Client.Post(request, &resp); err != nil {
		return nil, err
	}

	return resp.Data.User.Repositories.Nodes, nil
}
