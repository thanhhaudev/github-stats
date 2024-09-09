package github

var Queries = map[string]string{
	// repositoriesContributedTo: returns the repositories contributed to by the user
	// $username: the username of the user
	// $numRepos: the number of repositories to return
	// $afterCursor: the cursor to start from
	"repositories_contributed_to": `query ($username: String!, $numRepos: Int!, $afterCursor: String) {
	  user(login: $username) {
		repositoriesContributedTo(first: $numRepos, after: $afterCursor, orderBy: {field: CREATED_AT, direction: DESC}, includeUserRepositories: false) {
		  nodes {
			name
			url
			isPrivate
			isFork
			primaryLanguage {
				name
			}
            owner {
                login
            }
            languages(first: 10) {
                edges {
                    node {
                        name
                        color
                    }
                    size
                }
            }
		  }
		  pageInfo {
			endCursor
			hasNextPage
		  }
		}
	  }
	}`,
	// repositories: returns the repositories owned by the user
	// $username: the username of the user
	// $numRepos: the number of repositories to return
	// $afterCursor: the cursor to start from
	"repositories": `query ($username: String!, $numRepos: Int!, $afterCursor: String) {
	  user(login: $username) {
		repositories(first: $numRepos, after: $afterCursor, orderBy: {field: CREATED_AT, direction: DESC}, affiliations: [OWNER, COLLABORATOR], isFork: false) {
			nodes {
				name
				url
				isPrivate
				isFork
				primaryLanguage {
					name
				}
				owner {
					login
				}
				languages(first: 10) {
					edges {
						node {
							name
							color
						}
						size
					}
				}
			}
			pageInfo {
				endCursor
				hasNextPage
			}
		}
	  }
	}`,
	// repository_branches: returns the branches of a repository
	// $owner: the owner of the repository
	// $name: the name of the repository
	// $numBranches: the number of repositories to return
	// $afterCursor: the cursor to start from
	"repository_branches": `query ($owner: String!, $name: String!, $numBranches: Int!, $afterCursor: String) {
		repository(owner: $owner, name: $name) {
			refs(refPrefix: "refs/heads/", first: $numBranches, after: $afterCursor) {
				nodes {
					name
				}
				pageInfo {
					endCursor
					hasNextPage
				}
			}
		}
	}`,
	"repository_default_branch": `query ($owner: String!, $name: String!) {
		repository(owner: $owner, name: $name) {
			defaultBranchRef {
			  name
			}
		}
	}`,
	// repository_commits: returns the commits of a repository
	// $owner: the owner of the repository
	// $name: the name of the repository
	// $authorId: the ID of the author, e.g. "MDQ6VXNlcjc2OTQyMDAy"
	// $branch: the branch of the repository, e.g. "refs/heads/develop"
	// $perPage: the number of commits to return per page
	// $afterCursor: the cursor to start from
	"repository_commits": `query ($owner: String!, $name: String!, $authorId: ID!, $branch: String!, $numCommits: Int!, $afterCursor: String) {
		repository(owner: $owner, name: $name) {
			ref(qualifiedName: $branch) {
				target {
					... on Commit {
						history(author: { id: $authorId }, first: $numCommits, after: $afterCursor) {
							nodes {
								additions
								deletions
								committedDate
								oid
							}
							pageInfo {
								endCursor
								hasNextPage
							}
						}
					}
				}
			}
		}
	}`,
	// viewer: returns the viewer's information
	"viewer": `query {
	  viewer {
		id
		login
		name
		email
		createdAt
	  }
	}`,
}

type GitHub struct {
	Repositories *RepositoryService
	Viewer       *ViewerService
}

type PageInfo struct {
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}

type Language struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// NewGitHub creates a new GitHub
func NewGitHub(token string) *GitHub {
	if token == "" {
		return nil
	}

	client := NewClient(token)

	return &GitHub{
		Repositories: &RepositoryService{client},
		Viewer:       &ViewerService{client},
	}
}
