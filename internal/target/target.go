package target

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rhysd/actionlint"
	"github.com/rin2yh/gh-assert/internal/github"
	"github.com/rin2yh/gh-assert/internal/model"
)

const contractSuffix = "_assert.yml"

func Classify(contractPath string) (model.ContractKind, string, error) {
	if !strings.HasSuffix(filepath.Base(contractPath), contractSuffix) {
		return "", "", fmt.Errorf("%s: contract must be named <name>_assert.yml", contractPath)
	}

	siblingPath := strings.TrimSuffix(contractPath, contractSuffix) + ".yml"
	workflow := isWorkflow(siblingPath)
	composite := filepath.Base(contractPath) == "action_assert.yml" && isCompositeAction(siblingPath)
	kind, err := classify(workflow, composite)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", contractPath, err)
	}
	return kind, siblingPath, nil
}

func classify(workflow, composite bool) (model.ContractKind, error) {
	switch {
	case workflow && composite:
		return "", fmt.Errorf("sibling is both a Workflow and a Composite Action")
	case workflow:
		return model.WorkflowContract, nil
	case composite:
		return model.CompositeActionContract, nil
	default:
		return "", fmt.Errorf("sibling is neither a Workflow nor a Composite Action")
	}
}

func isWorkflow(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	_, parseErrors := actionlint.Parse(data)
	return len(parseErrors) == 0
}

func isCompositeAction(path string) bool {
	action, err := github.NewActionParser(path).Parse()
	return err == nil && action.Runs.Using == "composite"
}
