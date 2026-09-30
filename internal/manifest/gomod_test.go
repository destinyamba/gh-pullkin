package manifest

import (
	"testing"
)

func TestParseGoMod(t *testing.T) {
	data := []byte(`module example.com/test

go 1.26.0

require (
	github.com/a/one v1.0.0
	github.com/cli/go-gh/v2 v2.16.1 // indirect
	golang.org/x/mod v0.41.0
)
`)

	deps, err := ParseGoMod(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(deps) != 3 {
		t.Fatalf("got %d deps, want %d", len(deps), 3)
	}

	if deps[0].Name != "github.com/a/one" {
		t.Errorf("got name %q, want %q", deps[0].Name, "github.com/a/one")
	}

	if deps[1].Version != "v2.16.1" {
		t.Errorf("got version %q, want %q", deps[1].Version, "v2.16.1")
	}

	if deps[1].Direct {
		t.Errorf("got direct %v, want %v", deps[1].Direct, false)
	}
}

func TestParseGoModInvalid(t *testing.T) {
	data := []byte(` example.com/test

go 1.26.0

require (
github.com/a/one v1.0.0
	github.com/cli/go-gh/v2 v2.16.1 // indirect
	golang.org/x/mod v0.41.0
)
`)

	_, err := ParseGoMod(data)
	if err == nil {
		t.Fatal("got nill error, want an error")
	}

}
