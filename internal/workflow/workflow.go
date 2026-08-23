package workflow

import (
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

type definition struct {
	On struct {
		WorkflowCall *struct {
			Inputs map[string]input `yaml:"inputs"`
		} `yaml:"workflow_call"`
	} `yaml:"on"`
}

// ValidateReusableInterface validates a contract against its sibling reusable
// workflow. The returned boolean reports whether the target is reusable.
func ValidateReusableInterface(item contract.Loaded) (bool, error) {
	path, ok := workflowPath(item.Path)
	if !ok {
		return false, nil
	}

	inputs, reusable, err := loadReusableInputs(path)
	if err != nil || !reusable {
		return reusable, err
	}
	return true, compareInputs(path, inputs, item.Contract.Inputs)
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

	var workflow definition
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return nil, false, fmt.Errorf("%s: %w", path, err)
	}
	if workflow.On.WorkflowCall == nil {
		return nil, false, nil
	}
	return workflow.On.WorkflowCall.Inputs, true, nil
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
	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("%s: reusable workflow interface does not match contract: %s", filepath.Clean(path), strings.Join(problems, "; "))
}

func workflowType(contractType string) string {
	if contractType == "integer" {
		return "number"
	}
	return contractType
}
