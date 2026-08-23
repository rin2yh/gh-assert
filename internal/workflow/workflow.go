package workflow

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rin2yh/gh-assert/internal/contract"
	"gopkg.in/yaml.v3"
)

type input struct {
	Required bool   `yaml:"required"`
	Type     string `yaml:"type"`
}

// ValidateReusableInterface validates a contract against its sibling reusable
// workflow. The returned boolean reports whether the target is reusable.
func ValidateReusableInterface(item contract.Loaded) (bool, error) {
	workflowPath, ok := workflowPath(item.Path)
	if !ok {
		return false, nil
	}

	inputs, reusable, err := loadReusableInputs(workflowPath)
	if err != nil || !reusable {
		return reusable, err
	}
	if err := compareInputs(workflowPath, inputs, item.Contract.Inputs); err != nil {
		return true, err
	}
	return true, nil
}

func workflowPath(contractPath string) (string, bool) {
	const suffix = "_assert.yml"
	if !strings.HasSuffix(contractPath, suffix) {
		return "", false
	}
	return strings.TrimSuffix(contractPath, suffix) + ".yml", true
}

func loadReusableInputs(path string) (map[string]input, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false, err
	}

	var document yaml.Node
	if err := yaml.NewDecoder(bytes.NewReader(data)).Decode(&document); err != nil {
		return nil, false, fmt.Errorf("%s: %w", path, err)
	}
	if len(document.Content) == 0 {
		return nil, false, nil
	}

	on := mappingValue(document.Content[0], "on")
	call, reusable := eventNode(on, "workflow_call")
	if !reusable {
		return nil, false, nil
	}
	inputsNode := mappingValue(call, "inputs")
	if inputsNode == nil {
		return map[string]input{}, true, nil
	}

	var inputs map[string]input
	if err := inputsNode.Decode(&inputs); err != nil {
		return nil, true, fmt.Errorf("%s:%d:%d: workflow_call.inputs: %w", path, inputsNode.Line, inputsNode.Column, err)
	}
	return inputs, true, nil
}

func eventNode(on *yaml.Node, name string) (*yaml.Node, bool) {
	if on == nil {
		return nil, false
	}
	switch on.Kind {
	case yaml.MappingNode:
		value := mappingValue(on, name)
		return value, value != nil
	case yaml.ScalarNode:
		return on, on.Value == name
	case yaml.SequenceNode:
		for _, event := range on.Content {
			if event.Value == name {
				return event, true
			}
		}
	}
	return nil, false
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func compareInputs(path string, workflowInputs map[string]input, contractInputs map[string]contract.Rule) error {
	var problems []string
	for _, name := range slices.Sorted(maps.Keys(workflowInputs)) {
		declared := workflowInputs[name]
		rule, ok := contractInputs[name]
		if !ok {
			problems = append(problems, fmt.Sprintf("input %s has no contract", name))
			continue
		}
		if declared.Required != rule.Required {
			problems = append(problems, fmt.Sprintf("input %s required is %t in workflow_call and %t in contract", name, declared.Required, rule.Required))
		}
		if declared.Type != workflowType(rule.Type.Kind) {
			problems = append(problems, fmt.Sprintf("input %s type is %s in workflow_call and %s in contract", name, declared.Type, rule.Type.Kind))
		}
	}
	for _, name := range slices.Sorted(maps.Keys(contractInputs)) {
		if _, ok := workflowInputs[name]; !ok {
			problems = append(problems, fmt.Sprintf("contract input %s is not declared by workflow_call", name))
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("%s: reusable workflow interface does not match contract: %s", filepath.Clean(path), strings.Join(problems, "; "))
	}
	return nil
}

func workflowType(contractType string) string {
	if contractType == "integer" {
		return "number"
	}
	return contractType
}
