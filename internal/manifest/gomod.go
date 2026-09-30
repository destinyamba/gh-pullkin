package manifest

import (
	"errors"

	"golang.org/x/mod/modfile"
)

func ParseGoMod(data []byte) ([]Dep, error) {
	f, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return nil, errors.New("error parsing go.mod file")
	}

	var deps []Dep

	for _, r := range f.Require {
		deps = append(deps, Dep{
			Ecosystem: Go,
			Name:      r.Mod.Path,
			Version:   r.Mod.Version,
			Direct:    !r.Indirect,
		})
	}

	return deps, nil
}
