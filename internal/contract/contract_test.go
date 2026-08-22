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
)

func TestParseValidContract(t *testing.T) {
	path := filepath.Join("testdata", "valid.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	c, err := parse(path, data)
	if err != nil {
		t.Fatal(err)
	}

	got := fmt.Sprintf("env: %s\ninputs: %s", ruleNames(c.Env), ruleNames(c.Inputs))
	assertGolden(t, "valid.golden", got)
}

func TestParseReportsSectionInTypeErrors(t *testing.T) {
	path := filepath.Join("testdata", "inputs-missing-type.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, parseErr := parse(path, data)
	if parseErr == nil {
		t.Fatal("expected error")
	}
	assertGolden(t, "inputs-missing-type.golden", parseErr.Error())
}

func TestParseReportsSchemaErrors(t *testing.T) {
	path := filepath.Join("testdata", "schema-error.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, parseErr := parse(path, data)
	if parseErr == nil {
		t.Fatal("expected error")
	}
	assertGolden(t, "schema-error.golden", parseErr.Error())
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
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", tt.file+".yml")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := parse(path, data); err == nil {
				t.Fatal("expected error")
			}
		})
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

func ruleNames(rules map[string]Rule) string {
	return strings.Join(slices.Sorted(maps.Keys(rules)), ",")
}

func assertGolden(t *testing.T, filename, got string) {
	t.Helper()
	golden, err := os.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatal(err)
	}
	gotText := strings.TrimSpace(got)
	wantText := strings.TrimSpace(string(golden))
	if diff := cmp.Diff(wantText, gotText); diff != "" {
		t.Fatalf("golden mismatch (-want +got):\n%s", diff)
	}
}

func TestLoadTargetsDiscoversAssertFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "deploy_assert.yml"), []byte("env: {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "deploy.yml"), []byte("on: push\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadTargets(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 1 || filepath.Base(loaded[0].Path) != "deploy_assert.yml" {
		t.Fatalf("unexpected targets: %#v", loaded)
	}
}

func TestLoadTargetsRejectsEmptyDirectory(t *testing.T) {
	if _, err := LoadTargets(t.TempDir()); err == nil {
		t.Fatal("expected an error")
	}
}
