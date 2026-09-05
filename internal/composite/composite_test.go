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

	if err := Validate(compositeContractFile(contractPath, c)); err != nil {
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

			err = Validate(compositeContractFile(contractPath, c))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func fixtureContractPath(name string) string {
	return filepath.Join("testdata", name, contractName)
}

func compositeContractFile(path string, c *model.Contract) model.ContractFile {
	return model.ContractFile{
		Path:        path,
		SiblingPath: filepath.Join(filepath.Dir(path), "action.yml"),
		Kind:        model.CompositeActionContract,
		Contract:    c,
	}
}
