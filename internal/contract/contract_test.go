package contract

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rin2yh/gh-assert/internal/model"
	"github.com/rin2yh/gh-assert/internal/test"
)

func TestParseValidContract(t *testing.T) {
	c, err := parseFile(t, "valid")
	if err != nil {
		t.Fatal(err)
	}

	got := fmt.Sprintf("env: %s\ninputs: %s", ruleNames(c.Env), ruleNames(c.Inputs))
	assertGolden(t, "valid.golden", got)
}

func TestParseReportsDiagnostics(t *testing.T) {
	tests := []struct{ name, file string }{
		{name: "integer range", file: "schema-error"},
		{name: "missing type names its section", file: "inputs-missing-type"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertGolden(t, tt.file+".golden", assertParseFails(t, tt.file))
		})
	}
}

func TestParseRejectsUnsupportedDefinitions(t *testing.T) {
	tests := []struct {
		name string
		file string
	}{
		{name: "unknown field", file: "unknown-field"},
		{name: "missing type", file: "missing-type"},
		{name: "multiple types", file: "multiple-types"},
		{name: "invalid pattern", file: "bad-pattern"},
		{name: "unsupported default", file: "unsupported-default"},
		{name: "unsupported integer enum", file: "unsupported-integer-enum"},
		{name: "no sections", file: "no-sections"},
		{name: "inputs missing type", file: "inputs-missing-type"},
		{name: "inputs invalid pattern", file: "inputs-bad-pattern"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assertParseFails(t, tt.file) })
	}
}

func TestParseAcceptsSectionOnlyContracts(t *testing.T) {
	tests := []struct {
		name    string
		content string
		rules   func(*model.Contract) map[string]model.Rule
	}{
		{name: "env only", content: "env: {}\n", rules: func(c *model.Contract) map[string]model.Rule { return c.Env }},
		{name: "inputs only", content: "inputs: {}\n", rules: func(c *model.Contract) map[string]model.Rule { return c.Inputs }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := test.WriteFile(t, t.TempDir(), "contract.yml", tt.content)
			c, err := LoadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if tt.rules(c) == nil {
				t.Fatal("section was not parsed")
			}
		})
	}
}

func TestParseAcceptsEventOnlyContract(t *testing.T) {
	path := test.WriteFile(t, t.TempDir(), "contract.yml", "on:\n  workflow_run:\n    env: {}\n")
	c, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if c.On["workflow_run"].Env == nil {
		t.Fatal("event-specific env section was not parsed")
	}
}

func TestParseValidatesEventSpecificRules(t *testing.T) {
	path := test.WriteFile(t, t.TempDir(), "contract.yml", "on:\n  workflow_dispatch:\n    inputs:\n      retries:\n        type:\n          integer:\n            min: 2\n            max: 1\n")
	_, err := LoadFile(path)
	if err == nil || !strings.Contains(err.Error(), "min must not be greater than max") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadTargetsDiscoversAssertFiles(t *testing.T) {
	dir := workflowTestDir(t)
	test.WriteFile(t, dir, "deploy_assert.yml", "env: {}\n")
	test.WriteFile(t, dir, "deploy.yml", "on: push\n")
	loaded, err := LoadTargets(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 1 || filepath.Base(loaded[0].Path) != "deploy_assert.yml" {
		t.Fatalf("unexpected targets: %#v", loaded)
	}
}

func TestLoadTargetsRejectsWorkflowContractWithoutSibling(t *testing.T) {
	dir := workflowTestDir(t)
	path := test.WriteFile(t, dir, "deploy_assert.yml", "env: {}\n")

	for _, target := range []string{path, dir} {
		_, err := LoadTargets(target)
		if err == nil || !strings.Contains(err.Error(), "corresponding workflow") {
			t.Fatalf("LoadTargets(%q) error = %v", target, err)
		}
	}
}

func TestLoadTargetsRejectsNonContractFilename(t *testing.T) {
	path := test.WriteFile(t, workflowTestDir(t), "deploy_workflow_dispatch.yml", "env: {}\n")

	_, err := LoadTargets(path)
	if err == nil || !strings.Contains(err.Error(), "contract must be placed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadTargetsDoesNotTreatEventSuffixAsWorkflowName(t *testing.T) {
	dir := workflowTestDir(t)
	test.WriteFile(t, dir, "deploy.yml", "on: workflow_dispatch\n")
	path := test.WriteFile(t, dir, "deploy_workflow_dispatch_assert.yml", "env: {}\n")

	_, err := LoadTargets(path)
	if err == nil || !strings.Contains(err.Error(), "deploy_workflow_dispatch.yml") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadTargetsRejectsWorkflowContractOutsideWorkflowDirectory(t *testing.T) {
	dir := t.TempDir()
	test.WriteFile(t, dir, "deploy.yml", "on: push\n")
	path := test.WriteFile(t, dir, "deploy_assert.yml", "env: {}\n")

	_, err := LoadTargets(path)
	if err == nil || !strings.Contains(err.Error(), ".github/workflows") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadTargetsKeepsCompositeActionContractNaming(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ".github", "actions", "deploy")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := test.WriteFile(t, dir, "action_assert.yml", "env: {}\n")

	for _, target := range []string{path, root} {
		if _, err := LoadTargets(target); err != nil {
			t.Fatalf("LoadTargets(%q) error = %v", target, err)
		}
	}
}

func TestLoadTargetsRejectsActionContractFilenameOutsideActions(t *testing.T) {
	dir := t.TempDir()
	path := test.WriteFile(t, dir, "action_assert.yml", "env: {}\n")

	for _, target := range []string{path, dir} {
		_, err := LoadTargets(target)
		if err == nil || !strings.Contains(err.Error(), ".github/actions") {
			t.Fatalf("LoadTargets(%q) error = %v", target, err)
		}
	}
}

func TestLoadTargetsTreatsActionAssertInWorkflowsAsWorkflowContract(t *testing.T) {
	dir := workflowTestDir(t)
	test.WriteFile(t, dir, "action.yml", "on: push\n")
	path := test.WriteFile(t, dir, "action_assert.yml", "env: {}\n")

	if _, err := LoadTargets(path); err != nil {
		t.Fatal(err)
	}
}

func TestLoadTargetsRejectsEmptyDirectory(t *testing.T) {
	_, err := LoadTargets(t.TempDir())
	test.AssertError(t, err)
}

func workflowTestDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), ".github", "workflows")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func parseFile(t *testing.T, name string) (*model.Contract, error) {
	t.Helper()
	path := filepath.Join("testdata", name+".yml")
	return LoadFile(path)
}

func assertParseFails(t *testing.T, name string) string {
	t.Helper()
	_, err := parseFile(t, name)
	test.AssertError(t, err)
	return err.Error()
}

func ruleNames(rules map[string]model.Rule) string {
	return strings.Join(slices.Sorted(maps.Keys(rules)), ",")
}

func assertGolden(t *testing.T, filename, got string) {
	t.Helper()
	golden, err := os.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(strings.TrimSpace(string(golden)), strings.TrimSpace(got)); diff != "" {
		t.Fatalf("golden mismatch (-want +got):\n%s", diff)
	}
}
