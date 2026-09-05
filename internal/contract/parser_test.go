package contract

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestParserParse(t *testing.T) {
	path := filepath.Join("testdata", "contract.yml")
	parsed, err := NewParser(path).Parse()
	if err != nil {
		t.Fatal(err)
	}
	token := parsed.Env["TOKEN"]
	if !token.Required || token.Type.String == nil || len(token.Type.String.Enum) != 2 {
		t.Fatalf("unexpected env rule: %#v", token)
	}
	if token.Position.Path != path || token.Position.Line == 0 || token.TypePosition.Line == 0 {
		t.Fatalf("positions were not parsed: %#v", token)
	}
	retries := parsed.Inputs["retries"]
	if retries.Type.Integer == nil || retries.Type.Integer.Min == nil || *retries.Type.Integer.Min != 1 || retries.Type.Integer.Max == nil || *retries.Type.Integer.Max != 3 {
		t.Fatalf("unexpected input rule: %#v", retries)
	}
}

func TestParserParseEventSpecificRules(t *testing.T) {
	path := filepath.Join("testdata", "event-specific.yml")
	parsed, err := NewParser(path).Parse()
	if err != nil {
		t.Fatal(err)
	}
	rule := parsed.On["workflow_dispatch"].Inputs["deploy_type"]
	if !rule.Required || rule.Type.String == nil || len(rule.Type.String.Enum) != 2 {
		t.Fatalf("unexpected event-specific rule: %#v", rule)
	}
	if rule.Position.Path != path || rule.Position.Line == 0 || rule.TypePosition.Line == 0 {
		t.Fatalf("event-specific positions were not parsed: %#v", rule)
	}
}

func TestParserRejectsUnknownField(t *testing.T) {
	path := filepath.Join("testdata", "unknown_contract.yml")
	_, err := NewParser(path).Parse()
	if err == nil || !strings.Contains(err.Error(), "field unknown not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParserReturnsReadError(t *testing.T) {
	_, err := NewParser(filepath.Join(t.TempDir(), "missing.yml")).Parse()
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestContractLifecycle(t *testing.T) {
	parsed, err := NewParser("contract.yml").parse([]byte("env:\n  TOKEN:\n    type:\n      string:\n        pattern: '^token-'\n"))
	if err != nil {
		t.Fatal(err)
	}
	rule := parsed.Env["TOKEN"]
	if rule.Type.Kind != "" || rule.Type.String.Pattern != nil {
		t.Fatalf("parser prepared a rule: %#v", rule)
	}

	if err := Validate("contract.yml", parsed); err != nil {
		t.Fatal(err)
	}
	rule = parsed.Env["TOKEN"]
	if rule.Type.Kind != "" || rule.Type.String.Pattern != nil {
		t.Fatalf("validation prepared a rule: %#v", rule)
	}

	Compile(parsed)
	rule = parsed.Env["TOKEN"]
	if rule.Type.Kind != "string" || rule.Type.String.Pattern == nil {
		t.Fatalf("compiler did not prepare a rule: %#v", rule)
	}
}
