package composite

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

const contractName = "action_assert.yml"

// Validate compares a Composite Action's public input interface with its
// sibling contract. Contracts without a Composite Action sibling are ignored.
func Validate(item model.ContractFile) error {
	action, path, err := load(item.Path)
	if err != nil || action == nil || action.Runs.Using != "composite" {
		return err
	}
	return compareInputs(path, action.Inputs, item.Contract.Inputs)
}

// Is reports whether the contract belongs to a sibling Composite Action.
func Is(contractPath string) (bool, error) {
	action, _, err := load(contractPath)
	if err != nil || action == nil {
		return false, err
	}
	return action.Runs.Using == "composite", nil
}

func load(contractPath string) (*github.Action, string, error) {
	path, ok := actionPath(contractPath)
	if !ok {
		return nil, "", nil
	}
	action, err := github.NewActionParser(path).Parse()
	if os.IsNotExist(err) {
		return nil, path, nil
	}
	return action, path, err
}

func actionPath(contractPath string) (string, bool) {
	dir := filepath.Dir(contractPath)
	if filepath.Base(contractPath) != contractName ||
		filepath.Base(dir) == "workflows" && filepath.Base(filepath.Dir(dir)) == ".github" {
		return "", false
	}
	return filepath.Join(dir, "action.yml"), true
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
