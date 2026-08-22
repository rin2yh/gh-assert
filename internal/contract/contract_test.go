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
	"github.com/rin2yh/gh-assert/internal/testutil"
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
		rules   func(*Contract) map[string]Rule
	}{
		{name: "env only", content: "env: {}\n", rules: func(c *Contract) map[string]Rule { return c.Env }},
		{name: "inputs only", content: "inputs: {}\n", rules: func(c *Contract) map[string]Rule { return c.Inputs }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := parse("contract.yml", []byte(tt.content))
			if err != nil {
				t.Fatal(err)
			}
			if tt.rules(c) == nil {
				t.Fatal("section was not parsed")
			}
		})
	}
}

func TestLoadTargetsDiscoversAssertFiles(t *testing.T) {
	dir := t.TempDir()
	testutil.WriteFile(t, dir, "deploy_assert.yml", "env: {}\n")
	testutil.WriteFile(t, dir, "deploy.yml", "on: push\n")
	loaded, err := LoadTargets(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 1 || filepath.Base(loaded[0].Path) != "deploy_assert.yml" {
		t.Fatalf("unexpected targets: %#v", loaded)
	}
}

func TestLoadTargetsRejectsEmptyDirectory(t *testing.T) {
	_, err := LoadTargets(t.TempDir())
	testutil.AssertError(t, err)
}

func parseFile(t *testing.T, name string) (*Contract, error) {
	t.Helper()
	path := filepath.Join("testdata", name+".yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return parse(path, data)
}

func assertParseFails(t *testing.T, name string) string {
	t.Helper()
	_, err := parseFile(t, name)
	testutil.AssertError(t, err)
	return err.Error()
}

func ruleNames(rules map[string]Rule) string {
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
