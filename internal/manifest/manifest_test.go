package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScan(t *testing.T) {
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
