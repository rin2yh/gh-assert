package contract

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	SectionEnv    = "env"
	SectionInputs = "inputs"
)

var sections = []string{SectionEnv, SectionInputs}

type Position struct {
	Path   string
	Line   int
	Column int
}

type Contract struct {
	Env    map[string]Rule `yaml:"env"`
	Inputs map[string]Rule `yaml:"inputs"`
}

type Rule struct {
	Required bool     `yaml:"required,omitempty"`
	Type     Type     `yaml:"type"`
	Position Position `yaml:"-"`
}

type Type struct {
	String  *StringType  `yaml:"string,omitempty"`
	Integer *IntegerType `yaml:"integer,omitempty"`
	Boolean *struct{}    `yaml:"boolean,omitempty"`
	Kind    string       `yaml:"-"`
}

type StringType struct {
	Enum        []string       `yaml:"enum,omitempty"`
	PatternText string         `yaml:"pattern,omitempty"`
	Pattern     *regexp.Regexp `yaml:"-"`
}

type IntegerType struct {
	Min *int64 `yaml:"min,omitempty"`
	Max *int64 `yaml:"max,omitempty"`
}

type Loaded struct {
	Path     string
	Contract *Contract
}

func (c *Contract) Rules(section string) map[string]Rule {
	switch section {
	case SectionEnv:
		return c.Env
	case SectionInputs:
		return c.Inputs
	}
	return nil
}

func LoadFile(path string) (*Contract, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parse(path, data)
}

func LoadTargets(path string) ([]Loaded, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		c, err := LoadFile(path)
		if err != nil {
			return nil, err
		}
		return []Loaded{{Path: path, Contract: c}}, nil
	}

	var paths []string
	err = filepath.WalkDir(path, func(current string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_assert.yml") {
			paths = append(paths, current)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		return nil, fmt.Errorf("no contract files matching *_assert.yml found in %s", path)
	}

	loaded := make([]Loaded, 0, len(paths))
	for _, path := range paths {
		c, err := LoadFile(path)
		if err != nil {
			return nil, err
		}
		loaded = append(loaded, Loaded{Path: path, Contract: c})
	}
	return loaded, nil
}

func parse(path string, data []byte) (*Contract, error) {
	var c Contract
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&c); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if c.Env == nil && c.Inputs == nil {
		return nil, fmt.Errorf("%s: env or inputs is required", path)
	}

	positions := readPositions(path, data)
	for _, section := range sections {
		if err := prepareSection(section, c.Rules(section), positions[section]); err != nil {
			return nil, err
		}
	}
	return &c, nil
}

func prepareSection(section string, rules map[string]Rule, positions sectionPositions) error {
	names := make([]string, 0, len(rules))
	for name := range rules {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		rule := rules[name]
		rule.Position = positions.rules[name]
		typePosition := positions.types[name]
		if typePosition.Path == "" {
			typePosition = rule.Position
		}
		if err := prepareRule(section, name, &rule, typePosition); err != nil {
			return err
		}
		rules[name] = rule
	}
	return nil
}

func prepareRule(section, name string, rule *Rule, typePosition Position) error {
	types := 0
	if rule.Type.String != nil {
		types++
		rule.Type.Kind = "string"
		if rule.Type.String.PatternText != "" {
			pattern, err := regexp.Compile(rule.Type.String.PatternText)
			if err != nil {
				return fmt.Errorf("%s:%d:%d: invalid pattern: %w", typePosition.Path, typePosition.Line, typePosition.Column, err)
			}
			rule.Type.String.Pattern = pattern
		}
	}
	if rule.Type.Integer != nil {
		types++
		rule.Type.Kind = "integer"
		integer := rule.Type.Integer
		if integer.Min != nil && integer.Max != nil && *integer.Min > *integer.Max {
			return fmt.Errorf("%s:%d:%d: min must not be greater than max", typePosition.Path, typePosition.Line, typePosition.Column)
		}
	}
	if rule.Type.Boolean != nil {
		types++
		rule.Type.Kind = "boolean"
	}
	if types != 1 {
		return fmt.Errorf("%s:%d:%d: %s.%s.type must specify exactly one type", typePosition.Path, typePosition.Line, typePosition.Column, section, name)
	}
	return nil
}

type sectionPositions struct {
	rules map[string]Position
	types map[string]Position
}

func readPositions(path string, data []byte) map[string]sectionPositions {
	index := map[string]sectionPositions{}
	for _, section := range sections {
		index[section] = sectionPositions{rules: map[string]Position{}, types: map[string]Position{}}
	}
	var doc yaml.Node
	if yaml.Unmarshal(data, &doc) != nil || len(doc.Content) == 0 {
		return index
	}
	for _, section := range sections {
		node := mappingValue(doc.Content[0], section)
		if node == nil {
			continue
		}
		positions := index[section]
		for i := 0; i+1 < len(node.Content); i += 2 {
			name, rule := node.Content[i].Value, node.Content[i+1]
			positions.rules[name] = Position{Path: path, Line: rule.Line, Column: rule.Column}
			typeNode := mappingValue(rule, "type")
			if typeNode != nil && len(typeNode.Content) >= 2 {
				spec := typeNode.Content[1]
				positions.types[name] = Position{Path: path, Line: spec.Line, Column: spec.Column}
			}
		}
	}
	return index
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0]
	}
	if node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}
