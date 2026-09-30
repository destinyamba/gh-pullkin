package depsdev

import (
	"context"
	"errors"
	"net/http"

	"github.com/destinyamba/gh-pullkin/internal/manifest"
)

const baseURL = "https://api.deps.dev/v3"

type Repo struct {
	Owner string
	Name  string
}

type Client struct {
	http *http.Client
}

func New() *Client {
	return &Client{http: http.DefaultClient}
}

func (c *Client) SourceRepo(ctx context.Context, dep manifest.Dep) (Repo, error) {
	return Repo{}, errors.New("not implemented")
}
