package github

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ActionParser struct {
	path string
}

func NewActionParser(path string) *ActionParser {
	return &ActionParser{path: path}
}

func (p *ActionParser) Parse() (*Action, error) {
	data, err := os.ReadFile(p.path)
	if err != nil {
		return nil, err
	}
	return p.parse(data)
}

func (p *ActionParser) parse(data []byte) (*Action, error) {
	action := &Action{}
	if err := yaml.Unmarshal(data, action); err != nil {
		return nil, fmt.Errorf("%s: %w", p.path, err)
	}
	return action, nil
}
