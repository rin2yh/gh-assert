package parser

import (
	"bytes"
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

type Position struct {
	Path   string
	Line   int
	Column int
}

type Contract struct {
	Env    map[string]Rule `yaml:"env"`
	Inputs map[string]Rule `yaml:"inputs"`
}

type LoadedContract struct {
	Path     string
	Contract *Contract
}

type Rule struct {
	Required     bool     `yaml:"required,omitempty"`
	Type         Type     `yaml:"type"`
	Position     Position `yaml:"-"`
	TypePosition Position `yaml:"-"`
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

func LoadContract(path string) (*Contract, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseContract(path, data)
}

func ParseContract(path string, data []byte) (*Contract, error) {
	var parsed Contract
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&parsed); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	setPositions(path, data, &parsed)
	return &parsed, nil
}

func setPositions(path string, data []byte, parsed *Contract) {
	var document yaml.Node
	if yaml.Unmarshal(data, &document) != nil || len(document.Content) == 0 {
		return
	}
	setSectionPositions(path, mappingValue(document.Content[0], "env"), parsed.Env)
	setSectionPositions(path, mappingValue(document.Content[0], "inputs"), parsed.Inputs)
}

func setSectionPositions(path string, section *yaml.Node, rules map[string]Rule) {
	if section == nil {
		return
	}
	for i := 0; i+1 < len(section.Content); i += 2 {
		name, node := section.Content[i].Value, section.Content[i+1]
		rule := rules[name]
		rule.Position = Position{Path: path, Line: node.Line, Column: node.Column}
		if typeNode := mappingValue(node, "type"); typeNode != nil && len(typeNode.Content) >= 2 {
			spec := typeNode.Content[1]
			rule.TypePosition = Position{Path: path, Line: spec.Line, Column: spec.Column}
		}
		rules[name] = rule
	}
}
