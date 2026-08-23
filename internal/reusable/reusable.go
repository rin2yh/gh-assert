package reusable

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rin2yh/gh-assert/internal/github"
	"github.com/rin2yh/gh-assert/internal/model"
)

// Validate compares a reusable workflow's public input interface with its
// sibling contract. Contracts without a reusable sibling are ignored.
func Validate(item model.ContractFile) error {
	workflow, path, err := load(item.Path)
	if err != nil || workflow == nil {
		return err
	}
	call, reusable := workflow.Events["workflow_call"]
	if !reusable {
		return nil
	}
	return compareInputs(path, call.Inputs, item.Contract.Inputs)
}

func load(contractPath string) (*github.Workflow, string, error) {
	path, ok := workflowPath(contractPath)
	if !ok {
		return nil, "", nil
	}
	workflow, err := github.NewParser(path).Parse()
	if os.IsNotExist(err) {
		return nil, path, nil
	}
	return workflow, path, err
}

func workflowPath(contractPath string) (string, bool) {
	const suffix = "_assert.yml"
	if !strings.HasSuffix(contractPath, suffix) {
		return "", false
	}
	return strings.TrimSuffix(contractPath, suffix) + ".yml", true
}

func compareInputs(path string, workflowInputs map[string]github.WorkflowInput, contractInputs map[string]model.Rule) error {
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
