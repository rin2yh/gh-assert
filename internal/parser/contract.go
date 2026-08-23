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

	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil || len(document.Content) == 0 {
		return &parsed, nil
	}
	root := document.Content[0]
	return &model.Contract{
		Env:    rulesWithPositions(path, mappingValue(root, "env"), parsed.Env),
		Inputs: rulesWithPositions(path, mappingValue(root, "inputs"), parsed.Inputs),
	}, nil
}

func rulesWithPositions(path string, section *yaml.Node, rules map[string]model.Rule) map[string]model.Rule {
	if rules == nil {
		return nil
	}
	result := make(map[string]model.Rule, len(rules))
	for name, rule := range rules {
		result[name] = rule
	}
	if section == nil {
		return result
	}

	for i := 0; i+1 < len(section.Content); i += 2 {
		name, node := section.Content[i].Value, section.Content[i+1]
		rule, ok := result[name]
		if !ok {
			continue
		}
		rule.Position = model.Position{Path: path, Line: node.Line, Column: node.Column}
		if typeNode := mappingValue(node, "type"); typeNode != nil && len(typeNode.Content) >= 2 {
			spec := typeNode.Content[1]
			rule.TypePosition = model.Position{Path: path, Line: spec.Line, Column: spec.Column}
		}
		result[name] = rule
	}
	return result
}
