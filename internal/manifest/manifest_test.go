package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScan_GoMod(t *testing.T) {
	dir := t.TempDir()
	data := []byte(`module example.com/test

go 1.26.0

require (
	github.com/a/one v1.0.0
	github.com/cli/go-gh/v2 v2.16.1 // indirect
	golang.org/x/mod v0.41.0
)
`)

	err := os.WriteFile(filepath.Join(dir, "go.mod"), data, 0644)
	if err != nil {
		t.Fatalf("file: %v", err)
	}

	deps, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	if len(deps) != 3 {
		t.Errorf("expected 3 dependencies, found %d", len(deps))
	}
}

func TestScanNoGoMod(t *testing.T) {
	dir := t.TempDir()

	deps, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	if len(deps) != 0 {
		t.Errorf("expected 0 dependencies, found %d", len(deps))
	}
}

func TestScan_PackageJSON(t *testing.T) {
	dir := t.TempDir()
	data := []byte(`{
	"name": "test",
	"version": "1.0.0",
	"dependencies": {
		"express": "^4.18.2"
	}
}`)

	err := os.WriteFile(filepath.Join(dir, "package.json"), data, 0644)
	if err != nil {
		t.Fatalf("file: %v", err)
	}

	deps, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	if len(deps) != 1 {
		t.Fatalf("expected 1 dependency, found %d", len(deps))
	}

	got := deps[0]
	if got.Ecosystem != NPM {
		t.Errorf("expected ecosystem %q, found %q", NPM, got.Ecosystem)
	}
	if got.Name != "express" {
		t.Errorf("expected name %q, found %q", "express", got.Name)
	}
	if got.Version != "^4.18.2" {
		t.Errorf("expected version %q, found %q", "^4.18.2", got.Version)
	}
	if !got.Direct {
		t.Errorf("expected direct dependency")
	}
}

func TestScan_GoModAndPackageJSON(t *testing.T) {
	dir := t.TempDir()
	goMod := []byte(`module example.com/test

go 1.26.0

require (
	github.com/a/one v1.0.0
	github.com/b/two v1.2.0 // indirect
)
`)
	pkgJSON := []byte(`{
	"dependencies": {
		"express": "^4.18.2"
	}
}`)

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), goMod, 0644); err != nil {
		t.Fatalf("file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "package.json"), pkgJSON, 0644); err != nil {
		t.Fatalf("file: %v", err)
	}

	deps, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	if len(deps) != 3 {
		t.Fatalf("expected 3 dependencies, found %d", len(deps))
	}

	counts := map[Ecosystem]int{}
	for _, d := range deps {
		counts[d.Ecosystem]++
	}
	if counts[Go] != 2 {
		t.Errorf("expected 2 go dependencies, found %d", counts[Go])
	}
	if counts[NPM] != 1 {
		t.Errorf("expected 1 npm dependency, found %d", counts[NPM])
	}
}
