package workflow

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/rin2yh/gh-assert/internal/contract"
	"github.com/rin2yh/gh-assert/internal/test"
)

func TestValidateReusableInterface(t *testing.T) {
	dir := t.TempDir()
	workflowPath := test.WriteFile(t, dir, "deploy.yml", "on:\n  workflow_call:\n    inputs:\n      environment:\n        required: true\n        type: string\n      retries:\n        type: number\n")
	contractPath := test.WriteFile(t, dir, "deploy_assert.yml", "inputs:\n  environment:\n    required: true\n    type:\n      string: {}\n  retries:\n    type:\n      integer: {}\n")
	c, err := contract.LoadFile(contractPath)
	if err != nil {
		t.Fatal(err)
	}

	reusable, err := ValidateReusableInterface(contract.Loaded{Path: contractPath, Contract: c})
	if err != nil {
		t.Fatal(err)
	}
	if !reusable {
		t.Fatalf("%s was not recognized as reusable", workflowPath)
	}
}

func TestValidateReusableInterfaceRejectsMismatch(t *testing.T) {
	tests := []struct {
		name     string
		workflow string
		contract string
		want     string
	}{
		{name: "missing contract", workflow: "      environment:\n        type: string\n", contract: "inputs: {}\n", want: "environment has no contract"},
		{name: "missing workflow input", workflow: "      environment:\n        type: string\n", contract: "inputs:\n  environment:\n    type:\n      string: {}\n  extra:\n    type:\n      string: {}\n", want: "contract input extra"},
		{name: "required", workflow: "      environment:\n        required: true\n        type: string\n", contract: "inputs:\n  environment:\n    type:\n      string: {}\n", want: "required is true"},
		{name: "type", workflow: "      environment:\n        type: boolean\n", contract: "inputs:\n  environment:\n    type:\n      string: {}\n", want: "type is boolean"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			test.WriteFile(t, dir, "deploy.yml", "on:\n  workflow_call:\n    inputs:\n"+tt.workflow)
			contractPath := test.WriteFile(t, dir, "deploy_assert.yml", tt.contract)
			c, err := contract.LoadFile(contractPath)
			if err != nil {
				t.Fatal(err)
			}

			_, err = ValidateReusableInterface(contract.Loaded{Path: contractPath, Contract: c})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestValidateReusableInterfaceIgnoresRegularWorkflow(t *testing.T) {
	dir := t.TempDir()
	test.WriteFile(t, dir, "deploy.yml", "on:\n  workflow_dispatch:\n")
	contractPath := test.WriteFile(t, dir, "deploy_assert.yml", "inputs: {}\n")
	c, err := contract.LoadFile(contractPath)
	if err != nil {
		t.Fatal(err)
	}

	reusable, err := ValidateReusableInterface(contract.Loaded{Path: contractPath, Contract: c})
	if err != nil {
		t.Fatal(err)
	}
	if reusable {
		t.Fatal("regular workflow was recognized as reusable")
	}
}

func TestWorkflowPath(t *testing.T) {
	got, ok := workflowPath(filepath.Join(".github", "workflows", "deploy_assert.yml"))
	if !ok || got != filepath.Join(".github", "workflows", "deploy.yml") {
		t.Fatalf("workflowPath() = %q, %t", got, ok)
	}
}
