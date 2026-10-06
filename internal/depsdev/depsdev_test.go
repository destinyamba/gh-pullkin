package depsdev

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/destinyamba/gh-pullkin/internal/manifest"
)

type route struct {
	status int
	body   string
}

func fakeServer(t *testing.T, routes map[string]route) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rt, ok := routes[r.URL.EscapedPath()]
		if !ok {
			t.Errorf("unexpected request: %s", r.URL.EscapedPath())
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(rt.status)
		w.Write([]byte(rt.body))
	}))
	t.Cleanup(srv.Close)
	return &Client{http: srv.Client(), base: srv.URL}
}

func TestSourceRepo(t *testing.T) {
	c := fakeServer(t, map[string]route{
		"/systems/go/packages/github.com%2Fcli%2Fgo-gh%2Fv2/versions/v2.16.1": {http.StatusOK, `{
			"relatedProjects": [
				{"projectKey": {"id": "github.com/cli/go-gh"}, "relationType": "ISSUE_TRACKER"},
				{"projectKey": {"id": "github.com/cli/go-gh"}, "relationType": "SOURCE_REPO"}
			]
		}`},
	})

	dep := manifest.Dep{Ecosystem: manifest.Go, Name: "github.com/cli/go-gh/v2", Version: "v2.16.1"}
	repo, err := c.SourceRepo(context.Background(), dep)
	if err != nil {
		t.Fatalf("source repo: %v", err)
	}

	want := Repo{Owner: "cli", Name: "go-gh"}
	if repo != want {
		t.Errorf("expected %+v, found %+v", want, repo)
	}
}

func TestSourceRepoNPMUsesDefaultVersion(t *testing.T) {
	c := fakeServer(t, map[string]route{
		"/systems/npm/packages/express": {http.StatusOK, `{
			"versions": [
				{"versionKey": {"version": "4.18.2"}, "isDefault": false},
				{"versionKey": {"version": "5.2.1"}, "isDefault": true}
			]
		}`},
		"/systems/npm/packages/express/versions/5.2.1": {http.StatusOK, `{
			"relatedProjects": [
				{"projectKey": {"id": "github.com/expressjs/express"}, "relationType": "SOURCE_REPO"}
			]
		}`},
	})

	dep := manifest.Dep{Ecosystem: manifest.NPM, Name: "express", Version: "^4.18.2"}
	repo, err := c.SourceRepo(context.Background(), dep)
	if err != nil {
		t.Fatalf("source repo: %v", err)
	}

	want := Repo{Owner: "expressjs", Name: "express"}
	if repo != want {
		t.Errorf("expected %+v, found %+v", want, repo)
	}
}

func TestSourceRepoNPMNoDefault(t *testing.T) {
	c := fakeServer(t, map[string]route{
		"/systems/npm/packages/express": {http.StatusOK, `{
			"versions": [{"versionKey": {"version": "4.18.2"}, "isDefault": false}]
		}`},
	})

	dep := manifest.Dep{Ecosystem: manifest.NPM, Name: "express", Version: "^4.18.2"}
	_, err := c.SourceRepo(context.Background(), dep)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, found %v", err)
	}
}

func TestSourceRepoNotFound(t *testing.T) {
	c := fakeServer(t, map[string]route{
		"/systems/go/packages/github.com%2Fa%2Fb/versions/v9.9.9": {http.StatusNotFound, "version not found"},
	})

	dep := manifest.Dep{Ecosystem: manifest.Go, Name: "github.com/a/b", Version: "v9.9.9"}
	_, err := c.SourceRepo(context.Background(), dep)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, found %v", err)
	}
}

func TestSourceRepoNoGitHub(t *testing.T) {
	c := fakeServer(t, map[string]route{
		"/systems/go/packages/gitlab.com%2Fa%2Fb/versions/v1.0.0": {http.StatusOK, `{
			"relatedProjects": [
				{"projectKey": {"id": "gitlab.com/a/b"}, "relationType": "SOURCE_REPO"}
			]
		}`},
	})

	dep := manifest.Dep{Ecosystem: manifest.Go, Name: "gitlab.com/a/b", Version: "v1.0.0"}
	_, err := c.SourceRepo(context.Background(), dep)
	if !errors.Is(err, ErrNoRepo) {
		t.Errorf("expected ErrNoRepo, found %v", err)
	}
}

func TestSourceRepoServerError(t *testing.T) {
	c := fakeServer(t, map[string]route{
		"/systems/go/packages/github.com%2Fa%2Fb/versions/v1.0.0": {http.StatusInternalServerError, ""},
	})

	dep := manifest.Dep{Ecosystem: manifest.Go, Name: "github.com/a/b", Version: "v1.0.0"}
	_, err := c.SourceRepo(context.Background(), dep)
	if err == nil {
		t.Fatal("expected an error, found nil")
	}
}
