package manifest

import "errors"

type Ecosystem string

const (
	Go  Ecosystem = "go"
	NPM Ecosystem = "npm"
)

type Dep struct {
	Ecosystem Ecosystem
	Name      string
	Version   string
	Direct    bool
}

func Scan(dir string) ([]Dep, error) {
	return nil, errors.New("not implemented")
}
