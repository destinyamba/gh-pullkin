package issues

import (
	"context"
	"errors"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/destinyamba/gh-pullkin/internal/depsdev"
)

const MaxPerRepo = 20

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
	gh *api.RESTClient
}

func New() (*Client, error) {
	gh, err := api.DefaultRESTClient()
	if err != nil {
		return nil, err
	}
	return &Client{gh: gh}, nil
}

func (c *Client) Open(ctx context.Context, repo depsdev.Repo) ([]Issue, error) {
	return nil, errors.New("not implemented")
}

