package composite

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/rin2yh/gh-assert/internal/contract"
	"github.com/rin2yh/gh-assert/internal/model"
)

func TestValidateCompositeInterface(t *testing.T) {
	contractPath := fixtureContractPath("valid")
	c, err := contract.LoadFile(contractPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := Validate(model.ContractFile{Path: contractPath, Contract: c}); err != nil {
		t.Fatal(err)
	}
}

func TestValidateCompositeInterfaceRejectsMismatch(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "missing-contract", want: "environment has no contract"},
		{name: "missing-action-input", want: "contract input extra"},
		{name: "required-mismatch", want: "required is true"},
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

func TestValidateCompositeInterfaceIgnoresNonCompositeAction(t *testing.T) {
	contractPath := fixtureContractPath("javascript-action")
	c, err := contract.LoadFile(contractPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := Validate(model.ContractFile{Path: contractPath, Contract: c}); err != nil {
		t.Fatal(err)
	}
}

func TestValidateCompositeInterfaceAllowsMissingAction(t *testing.T) {
	contractPath := fixtureContractPath("missing-action")
	c, err := contract.LoadFile(contractPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := Validate(model.ContractFile{Path: contractPath, Contract: c}); err != nil {
		t.Fatal(err)
	}
}

func TestIs(t *testing.T) {
	composite, err := Is(fixtureContractPath("valid"))
	if err != nil {
		t.Fatal(err)
	}
	if !composite {
		t.Fatal("Is() = false, want true")
	}
}

func fixtureContractPath(name string) string {
	return filepath.Join("testdata", name, contractName)
}

func TestActionPath(t *testing.T) {
	got, ok := actionPath(filepath.Join("custom", "deploy", contractName))
	want := filepath.Join("custom", "deploy", "action.yml")
	if !ok || got != want {
		t.Fatalf("actionPath() = %q, %t, want %q, true", got, ok, want)
	}
	if _, ok := actionPath(filepath.Join(".github", "workflows", contractName)); ok {
		t.Fatal("actionPath() treated a Workflow contract as a Composite Action contract")
	}
}
