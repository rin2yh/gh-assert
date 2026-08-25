package target

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rin2yh/gh-assert/internal/model"
	"github.com/rin2yh/gh-assert/internal/test"
)

const workflow = "on: push\njobs:\n  test:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo test\n"

func TestClassify(t *testing.T) {
	tests := []struct {
		name, contractName, sibling string
		wantKind                    model.ContractKind
	}{
		{name: "workflow", contractName: "deploy_assert.yml", sibling: workflow, wantKind: model.WorkflowContract},
		{name: "workflow named action", contractName: "action_assert.yml", sibling: workflow, wantKind: model.WorkflowContract},
		{name: "composite action", contractName: "action_assert.yml", sibling: "name: test\nruns:\n  using: composite\n  steps: []\n", wantKind: model.CompositeActionContract},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), "custom")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			path := test.WriteFile(t, dir, tt.contractName, "env: {}\n")
			sibling := strings.TrimSuffix(tt.contractName, contractSuffix) + ".yml"
			test.WriteFile(t, dir, sibling, tt.sibling)

			kind, siblingPath, err := Classify(path)
			if err != nil {
				t.Fatal(err)
			}
			if kind != tt.wantKind || siblingPath != filepath.Join(dir, sibling) {
				t.Fatalf("Classify() = %q, %q, want %q, %q", kind, siblingPath, tt.wantKind, filepath.Join(dir, sibling))
			}
		})
	}
}

func TestClassifyRejectsUnknownTarget(t *testing.T) {
	tests := []struct {
		name, contract, sibling string
	}{
		{name: "missing sibling", contract: "deploy_assert.yml"},
		{name: "non-composite action", contract: "action_assert.yml", sibling: "name: test\nruns:\n  using: node20\n  main: index.js\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := test.WriteFile(t, dir, tt.contract, "env: {}\n")
			if tt.sibling != "" {
				test.WriteFile(t, dir, strings.TrimSuffix(tt.contract, contractSuffix)+".yml", tt.sibling)
			}
			_, _, err := Classify(path)
			if err == nil || !strings.Contains(err.Error(), "neither") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestClassifyRejectsAmbiguousTarget(t *testing.T) {
	_, err := classify(true, true)
	if err == nil || !strings.Contains(err.Error(), "both") {
		t.Fatalf("unexpected error: %v", err)
	}
}
