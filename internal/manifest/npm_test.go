package manifest

import "testing"

func TestParsePackageJSON(t *testing.T) {
	data := []byte(`{
		"dependencies": {
			"lodash": "^4.17.20"
		},
		"devDependencies": {
			"jest": "^26.6.3"
		}
	}`)

	deps, err := ParsePackageJSON(data)
	if err != nil {
		t.Fatalf("failed to parse package.json: %v", err)
	}

	if len(deps) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(deps))
	}

	if deps[0].Name != "lodash" || deps[0].Version != "^4.17.20" {
		t.Fatalf("expected lodash dependency")
	}

	if deps[1].Name != "jest" || deps[1].Version != "^26.6.3" {
		t.Fatalf("expected jest dependency")
	}

}

func TestParsePackageJSONInvalid(t *testing.T) {
	data := []byte(`invalid json`)
	_, err := ParsePackageJSON(data)
	if err == nil {
		t.Fatalf("expected error for invalid package.json")
	}
}
