package github

var Queries = map[string]string{
	// repositoriesContributedTo: returns the repositories contributed to by the user
	"repositoriesContributedTo": `query ($username: String!, $numRepos: Int!, $afterCursor: String) {
	  user(login: $username) {
		repositoriesContributedTo(first: $numRepos, after: $afterCursor, orderBy: {field: CREATED_AT, direction: DESC}, includeUserRepositories: true) {
		  nodes {
			name
			isPrivate
			isFork
			primaryLanguage {
			  name
			}
		  }
		  pageInfo {
			endCursor
			hasNextPage
		  }
		}
	  }
	}`,
}

type GitHub struct {
	Repositories *RepositoryService
}

// NewGitHub creates a new GitHub
func NewGitHub(token string) *GitHub {
	client := NewClient(token)

	return &GitHub{
		Repositories: &RepositoryService{client},
	}
}
