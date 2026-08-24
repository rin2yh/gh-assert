package reusable

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rhysd/actionlint"
	"github.com/rin2yh/gh-assert/internal/model"
)

// Validate compares a reusable workflow's public input interface with its
// sibling contract. Contracts without a reusable sibling are ignored.
func Validate(item model.ContractFile) error {
	workflow, path, err := load(item.Path)
	if err != nil || workflow == nil {
		return err
	}
	call, reusable := workflow.FindWorkflowCallEvent()
	if !reusable {
		return nil
	}
	contractInputs := item.Contract.Inputs
	scopedInputs := item.Contract.On["workflow_call"].Inputs
	if len(scopedInputs) > 0 {
		contractInputs = maps.Clone(contractInputs)
		if contractInputs == nil {
			contractInputs = make(map[string]model.Rule, len(scopedInputs))
		}
		maps.Copy(contractInputs, scopedInputs)
	}
	return compareInputs(path, call, contractInputs)
}

func load(contractPath string) (*actionlint.Workflow, string, error) {
	path, ok := workflowPath(contractPath)
	if !ok {
		return nil, "", nil
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, path, nil
	}
	if err != nil {
		return nil, path, err
	}
	workflow, parseErrors := actionlint.Parse(data)
	if len(parseErrors) > 0 {
		first := parseErrors[0]
		return nil, path, fmt.Errorf("%s:%d:%d: %s", path, first.Line, first.Column, first.Message)
	}
	return workflow, path, nil
}

func workflowPath(contractPath string) (string, bool) {
	const suffix = "_assert.yml"
	if !strings.HasSuffix(contractPath, suffix) || filepath.Base(contractPath) == "action_assert.yml" {
		return "", false
	}
	return strings.TrimSuffix(contractPath, suffix) + ".yml", true
}

func compareInputs(path string, call *actionlint.WorkflowCallEvent, contractInputs map[string]model.Rule) error {
	var problems []string
	for _, declared := range call.Inputs {
		name := declared.ID
		rule, ok := contractInputs[name]
		if !ok {
			problems = append(problems, fmt.Sprintf("input %s has no contract", name))
			continue
		}
		if declared.IsRequired() != rule.Required {
			problems = append(problems, fmt.Sprintf("input %s required is %t in workflow_call and %t in contract", name, declared.IsRequired(), rule.Required))
		}
		if declared.Type != workflowType(rule.Type.Kind) {
			problems = append(problems, fmt.Sprintf("input %s type is %s in workflow_call and %s in contract", name, workflowTypeName(declared.Type), rule.Type.Kind))
		}
	}
	for _, name := range slices.Sorted(maps.Keys(contractInputs)) {
		if !slices.ContainsFunc(call.Inputs, func(input *actionlint.WorkflowCallEventInput) bool { return input.ID == name }) {
			problems = append(problems, fmt.Sprintf("contract input %s is not declared by workflow_call", name))
		}
	}
	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("%s: reusable workflow interface does not match contract: %s", filepath.Clean(path), strings.Join(problems, "; "))
}

func workflowType(contractType string) actionlint.WorkflowCallEventInputType {
	switch contractType {
	case "boolean":
		return actionlint.WorkflowCallEventInputTypeBoolean
	case "integer":
		return actionlint.WorkflowCallEventInputTypeNumber
	case "string":
		return actionlint.WorkflowCallEventInputTypeString
	default:
		return actionlint.WorkflowCallEventInputTypeInvalid
	}
}

func workflowTypeName(inputType actionlint.WorkflowCallEventInputType) string {
	switch inputType {
	case actionlint.WorkflowCallEventInputTypeBoolean:
		return "boolean"
	case actionlint.WorkflowCallEventInputTypeNumber:
		return "number"
	case actionlint.WorkflowCallEventInputTypeString:
		return "string"
	default:
		return "invalid"
	}
}
