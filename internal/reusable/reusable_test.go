package reusable

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/rin2yh/gh-assert/internal/contract"
	"github.com/rin2yh/gh-assert/internal/model"
)

func TestValidateReusableInterface(t *testing.T) {
	contractPath := fixtureContractPath("valid")
	c, err := contract.LoadFile(contractPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := Validate(model.ContractFile{Path: contractPath, Contract: c}); err != nil {
		t.Fatal(err)
	}
}

func TestValidateReusableInterfaceUsesWorkflowCallRules(t *testing.T) {
	contractPath := fixtureContractPath("event-specific")
	c, err := contract.LoadFile(contractPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := Validate(model.ContractFile{Path: contractPath, Contract: c}); err != nil {
		t.Fatal(err)
	}
}

func TestValidateReusableInterfaceRejectsMismatch(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "missing-contract", want: "environment has no contract"},
		{name: "missing-workflow-input", want: "contract input extra"},
		{name: "required-mismatch", want: "required is true"},
		{name: "type-mismatch", want: "type is boolean"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contractPath := fixtureContractPath(tt.name)
			c, err := contract.LoadFile(contractPath)
			if err != nil {
				t.Fatal(err)
			}

			err = Validate(model.ContractFile{Path: contractPath, Contract: c})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestValidateReusableInterfaceIgnoresRegularWorkflow(t *testing.T) {
	contractPath := fixtureContractPath("regular-workflow")
	c, err := contract.LoadFile(contractPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := Validate(model.ContractFile{Path: contractPath, Contract: c}); err != nil {
		t.Fatal(err)
	}
}

func TestValidateReusableInterfaceAllowsMissingWorkflow(t *testing.T) {
	contractPath := fixtureContractPath("missing-workflow")
	c, err := contract.LoadFile(contractPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := Validate(model.ContractFile{Path: contractPath, Contract: c}); err != nil {
		t.Fatal(err)
	}
}

func fixtureContractPath(name string) string {
	return filepath.Join("testdata", name, ".github", "workflows", "deploy_assert.yml")
}

func TestWorkflowPath(t *testing.T) {
	for _, name := range []string{"deploy", "action"} {
		got, ok := workflowPath(filepath.Join(".github", "workflows", name+"_assert.yml"))
		if !ok || got != filepath.Join(".github", "workflows", name+".yml") {
			t.Fatalf("workflowPath() = %q, %t", got, ok)
		}
	}
	if _, ok := workflowPath(filepath.Join(".github", "actions", "deploy", "action_assert.yml")); ok {
		t.Fatal("workflowPath() treated a Composite Action contract as a Workflow contract")
	}
}
