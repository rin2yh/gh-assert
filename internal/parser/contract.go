package parser

import (
	"bytes"
	"fmt"
	"os"

	"github.com/rin2yh/gh-assert/internal/model"
	"gopkg.in/yaml.v3"
)

type ContractParser struct{}

func (ContractParser) Parse(path string) (*model.Contract, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseContract(path, data)
}

func parseContract(path string, data []byte) (*model.Contract, error) {
	var parsed model.Contract
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&parsed); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	setPositions(path, data, &parsed)
	return &parsed, nil
}

func setPositions(path string, data []byte, parsed *model.Contract) {
	var document yaml.Node
	if yaml.Unmarshal(data, &document) != nil || len(document.Content) == 0 {
		return
	}
	setSectionPositions(path, mappingValue(document.Content[0], "env"), parsed.Env)
	setSectionPositions(path, mappingValue(document.Content[0], "inputs"), parsed.Inputs)
}

func setSectionPositions(path string, section *yaml.Node, rules map[string]model.Rule) {
	if section == nil {
		return
	}
	for i := 0; i+1 < len(section.Content); i += 2 {
		name, node := section.Content[i].Value, section.Content[i+1]
		rule := rules[name]
		rule.Position = model.Position{Path: path, Line: node.Line, Column: node.Column}
		if typeNode := mappingValue(node, "type"); typeNode != nil && len(typeNode.Content) >= 2 {
			spec := typeNode.Content[1]
			rule.TypePosition = model.Position{Path: path, Line: spec.Line, Column: spec.Column}
		}
		rules[name] = rule
	}
}
