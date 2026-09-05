package composite

import (
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rin2yh/gh-assert/internal/github"
	"github.com/rin2yh/gh-assert/internal/model"
)

const contractName = "action_assert.yml"

// Validate compares a Composite Action's public input interface with its sibling contract.
func Validate(item model.ContractFile) error {
	if item.Kind != model.CompositeActionContract {
		return nil
	}
	action, err := load(item.SiblingPath)
	if err != nil {
		return err
	}
	return compareInputs(item.SiblingPath, action.Inputs, item.Contract.Inputs)
}

func load(path string) (*github.Action, error) {
	action, err := github.NewActionParser(path).Parse()
	return action, err
}

func compareInputs(path string, actionInputs map[string]github.ActionInput, contractInputs map[string]model.Rule) error {
	var problems []string
	for _, name := range slices.Sorted(maps.Keys(actionInputs)) {
		declared := actionInputs[name]
		rule, ok := contractInputs[name]
		if !ok {
			problems = append(problems, fmt.Sprintf("input %s has no contract", name))
			continue
		}
		if declared.Required != rule.Required {
			problems = append(problems, fmt.Sprintf("input %s required is %t in action.yml and %t in contract", name, declared.Required, rule.Required))
		}
	}
	for _, name := range slices.Sorted(maps.Keys(contractInputs)) {
		if _, ok := actionInputs[name]; !ok {
			problems = append(problems, fmt.Sprintf("contract input %s is not declared by action.yml", name))
		}
	}
	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("%s: composite action interface does not match contract: %s", filepath.Clean(path), strings.Join(problems, "; "))
}
