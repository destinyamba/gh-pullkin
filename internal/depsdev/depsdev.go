package depsdev

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/destinyamba/gh-pullkin/internal/manifest"
)

const baseURL = "https://api.deps.dev/v3"

var (
	ErrNotFound = errors.New("package version not found on deps.dev")
	ErrNoRepo   = errors.New("no GitHub source repo")
)

type Repo struct {
	Owner string
	Name  string
}

type Client struct {
	http *http.Client
	base string
}

func New() *Client {
	return &Client{http: http.DefaultClient, base: baseURL}
}

type packageResponse struct {
	Versions []struct {
		VersionKey struct {
			Version string `json:"version"`
		} `json:"versionKey"`
		IsDefault bool `json:"isDefault"`
	} `json:"versions"`
}

type versionResponse struct {
	RelatedProjects []struct {
		ProjectKey struct {
			ID string `json:"id"`
		} `json:"projectKey"`
		RelationType string `json:"relationType"`
	} `json:"relatedProjects"`
}

func (c *Client) SourceRepo(ctx context.Context, dep manifest.Dep) (Repo, error) {
	version := dep.Version
	if dep.Ecosystem == manifest.NPM {
		v, err := c.defaultVersion(ctx, dep)
		if err != nil {
			return Repo{}, err
		}
		version = v
	}

	var body versionResponse
	path := fmt.Sprintf("/systems/%s/packages/%s/versions/%s",
		dep.Ecosystem, url.PathEscape(dep.Name), url.PathEscape(version))
	if err := c.get(ctx, path, &body); err != nil {
		return Repo{}, fmt.Errorf("%s %s: %w", dep.Name, version, err)
	}

	for _, p := range body.RelatedProjects {
		if p.RelationType != "SOURCE_REPO" {
			continue
		}
		parts := strings.Split(p.ProjectKey.ID, "/")
		if len(parts) != 3 || parts[0] != "github.com" {
			continue
		}
		return Repo{Owner: parts[1], Name: parts[2]}, nil
	}

	return Repo{}, fmt.Errorf("%s: %w", dep.Name, ErrNoRepo)
}

func (c *Client) defaultVersion(ctx context.Context, dep manifest.Dep) (string, error) {
	var body packageResponse
	path := fmt.Sprintf("/systems/%s/packages/%s", dep.Ecosystem, url.PathEscape(dep.Name))
	if err := c.get(ctx, path, &body); err != nil {
		return "", fmt.Errorf("%s: %w", dep.Name, err)
	}

	for _, v := range body.Versions {
		if v.IsDefault {
			return v.VersionKey.Version, nil
		}
	}

	return "", fmt.Errorf("%s: no default version: %w", dep.Name, ErrNotFound)
}

func (c *Client) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+path, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("deps.dev: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("deps.dev: status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode deps.dev: %w", err)
	}
	return nil
}
