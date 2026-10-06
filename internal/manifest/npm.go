package manifest

import (
	"encoding/json"
	"fmt"
)

func ParsePackageJSON(data []byte) ([]Dep, error) {
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("parse package.json: %w", err)
	}

	deps := make([]Dep, 0, len(pkg.Dependencies)+len(pkg.DevDependencies))

	for name, version := range pkg.Dependencies {
		deps = append(deps, Dep{
			Ecosystem: NPM,
			Name:      name,
			Version:   version,
			Direct:    true,
		})
	}

	for name, version := range pkg.DevDependencies {
		deps = append(deps, Dep{
			Ecosystem: NPM,
			Name:      name,
			Version:   version,
			Direct:    true,
		})
	}

	return deps, nil
}
