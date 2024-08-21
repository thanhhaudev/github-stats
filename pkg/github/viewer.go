package github

type ViewerService struct {
	Client *Client
}

type Viewer struct {
	ID        string `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"createdAt"`
}

// Get returns the viewer's information
func (v *ViewerService) Get(request *Request) (*Viewer, error) {
	var resp struct {
		Data struct {
			Viewer *Viewer `json:"viewer"`
		} `json:"data"`
	}

	if err := v.Client.Post(request, &resp); err != nil {
		return nil, err
	}

	return resp.Data.Viewer, nil
}
