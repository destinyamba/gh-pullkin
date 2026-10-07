package issues

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/destinyamba/gh-pullkin/internal/depsdev"
)

type fakeTransport func(*http.Request) (*http.Response, error)

func (f fakeTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func fakeClient(t *testing.T, body string) *Client {
	t.Helper()
	gql, err := api.NewGraphQLClient(api.ClientOptions{
		Host:      "github.com",
		AuthToken: "test",
		Transport: fakeTransport(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": {"application/json"}},
				Body:       io.NopCloser(strings.NewReader(body)),
				Request:    r,
			}, nil
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	return &Client{gh: gql}
}

func TestOpen(t *testing.T) {
	c := fakeClient(t, `{"data":{"search":{"nodes":[
	{
		"number": 1,
		"title": "taken one",
		"url": "https://github.com/cli/cli/issues/1",
		"updatedAt": "2026-09-01T00:00:00Z",
		"labels": {"nodes": [{"name": "help wanted"}]},
		"assignees": {"nodes": [{"login": "octocat"}]},
		"timelineItems": {"nodes": [{"createdAt": "2026-08-01T00:00:00Z"}]},
		"closedByPullRequestsReferences": {"nodes": [{"state": "OPEN"}]}
	},
	{
		"number": 2,
		"title": "free one",
		"url": "https://github.com/cli/cli/issues/2",
		"updatedAt": "2026-09-02T00:00:00Z",
		"labels": {"nodes": []},
		"assignees": {"nodes": []},
		"timelineItems": {"nodes": []},
		"closedByPullRequestsReferences": {"nodes": []}
	}
]}}}`)

	iss, err := c.Open(context.Background(), depsdev.Repo{Owner: "cli", Name: "cli"})
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	if len(iss) != 2 {
		t.Fatalf("expected 2 issues, found %d", len(iss))
	}

	taken := iss[0]
	if !taken.HasOpenPR {
		t.Errorf("issue 1: expected an open PR")
	}
	if len(taken.Assignees) != 1 || taken.Assignees[0] != "octocat" {
		t.Errorf("issue 1: expected assignees [octocat], found %v", taken.Assignees)
	}
	wantAssigned := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	if !taken.AssignedAt.Equal(wantAssigned) {
		t.Errorf("issue 1: expected assigned at %v, found %v", wantAssigned, taken.AssignedAt)
	}

	free := iss[1]
	if free.HasOpenPR {
		t.Errorf("issue 2: expected no open PR")
	}
	if len(free.Assignees) != 0 {
		t.Errorf("issue 2: expected no assignees, found %v", free.Assignees)
	}
	if !free.AssignedAt.IsZero() {
		t.Errorf("issue 2: expected zero assigned at, found %v", free.AssignedAt)
	}
}

func TestInfo(t *testing.T) {
	c := fakeClient(t, `{"data":{"repository":{"isArchived":true,"pushedAt":"2026-01-01T00:00:00Z"}}}`)

	info, err := c.Info(context.Background(), depsdev.Repo{Owner: "cli", Name: "cli"})
	if err != nil {
		t.Fatalf("info: %v", err)
	}

	if !info.Archived {
		t.Errorf("expected archived")
	}

	want := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if !info.PushedAt.Equal(want) {
		t.Errorf("expected pushed at %v, found %v", want, info.PushedAt)
	}
}
