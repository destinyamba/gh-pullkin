package issues

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/destinyamba/gh-pullkin/internal/depsdev"
)

const MaxPerRepo = 20

const searchQuery = `
query($q: String!, $first: Int!) {
  search(query: $q, type: ISSUE, first: $first) {
    nodes {
      ... on Issue {
        number
        title
        url
        updatedAt
        labels(first: 10) { nodes { name } }
        assignees(first: 5) { nodes { login } }
        timelineItems(itemTypes: [ASSIGNED_EVENT], last: 1) {
          nodes { ... on AssignedEvent { createdAt } }
        }
        closedByPullRequestsReferences(first: 5, includeClosedPrs: false) {
          nodes { state }
        }
      }
    }
  }
}
`

const repoInfoQuery = `
query($owner: String!, $name: String!) {
  repository(owner: $owner, name: $name) {
    isArchived
    pushedAt
  }
}
`

type searchResponse struct {
	Search struct {
		Nodes []struct {
			Number    int       `json:"number"`
			Title     string    `json:"title"`
			URL       string    `json:"url"`
			UpdatedAt time.Time `json:"updatedAt"`
			Labels    struct {
				Nodes []struct {
					Name string `json:"name"`
				} `json:"nodes"`
			} `json:"labels"`
			Assignees struct {
				Nodes []struct {
					Login string `json:"login"`
				} `json:"nodes"`
			} `json:"assignees"`
			TimelineItems struct {
				Nodes []struct {
					CreatedAt time.Time `json:"createdAt"`
				} `json:"nodes"`
			} `json:"timelineItems"`
			ClosedByPullRequestsReferences struct {
				Nodes []struct {
					State string `json:"state"`
				} `json:"nodes"`
			} `json:"closedByPullRequestsReferences"`
		} `json:"nodes"`
	} `json:"search"`
}

type repoInfoResponse struct {
	Repository struct {
		IsArchived bool      `json:"isArchived"`
		PushedAt   time.Time `json:"pushedAt"`
	} `json:"repository"`
}

var Labels = []string{
	"good first issue",
	"good-first-issue",
	"help wanted",
	"help-wanted",
	"beginner",
	"beginner friendly",
	"easy",
	"first-timers-only",
	"up-for-grabs",
}

type Issue struct {
	Repo       depsdev.Repo
	Number     int
	Title      string
	URL        string
	Labels     []string
	Assignees  []string
	AssignedAt time.Time
	HasOpenPR  bool
	UpdatedAt  time.Time
}

type RepoInfo struct {
	Archived bool
	PushedAt time.Time
}

type Client struct {
	gh *api.GraphQLClient
}

func New() (*Client, error) {
	gh, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, err
	}
	return &Client{gh: gh}, nil
}

func (c *Client) Open(ctx context.Context, repo depsdev.Repo) ([]Issue, error) {
	quoted := make([]string, 0, len(Labels))
	for _, l := range Labels {
		quoted = append(quoted, fmt.Sprintf("%q", l))
	}
	search := fmt.Sprintf("repo:%s/%s is:issue is:open label:%s sort:updated-desc",
		repo.Owner, repo.Name, strings.Join(quoted, ","))

	var resp searchResponse
	vars := map[string]interface{}{"q": search, "first": MaxPerRepo}
	if err := c.gh.DoWithContext(ctx, searchQuery, vars, &resp); err != nil {
		return nil, fmt.Errorf("search issues %s/%s: %w", repo.Owner, repo.Name, err)
	}

	issues := make([]Issue, 0, len(resp.Search.Nodes))
	for _, node := range resp.Search.Nodes {
		labels := make([]string, 0, len(node.Labels.Nodes))
		for _, l := range node.Labels.Nodes {
			labels = append(labels, l.Name)
		}

		assignees := make([]string, 0, len(node.Assignees.Nodes))
		for _, a := range node.Assignees.Nodes {
			assignees = append(assignees, a.Login)
		}

		var assignedAt time.Time
		if len(node.TimelineItems.Nodes) > 0 {
			assignedAt = node.TimelineItems.Nodes[0].CreatedAt
		}

		issues = append(issues, Issue{
			Repo:       repo,
			Number:     node.Number,
			Title:      node.Title,
			URL:        node.URL,
			Labels:     labels,
			Assignees:  assignees,
			AssignedAt: assignedAt,
			HasOpenPR:  len(node.ClosedByPullRequestsReferences.Nodes) > 0,
			UpdatedAt:  node.UpdatedAt,
		})
	}
	return issues, nil
}

func (c *Client) Info(ctx context.Context, repo depsdev.Repo) (RepoInfo, error) {
	var resp repoInfoResponse
	vars := map[string]interface{}{"owner": repo.Owner, "name": repo.Name}
	if err := c.gh.DoWithContext(ctx, repoInfoQuery, vars, &resp); err != nil {
		return RepoInfo{}, fmt.Errorf("get repo info %s/%s: %w", repo.Owner, repo.Name, err)
	}

	return RepoInfo{
		Archived: resp.Repository.IsArchived,
		PushedAt: resp.Repository.PushedAt,
	}, nil
}
