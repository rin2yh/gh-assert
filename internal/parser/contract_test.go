package parser

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestContractParserParse(t *testing.T) {
	path := filepath.Join("testdata", "contract.yml")
	parsed, err := (ContractParser{}).Parse(path)
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

func TestContractParserRejectsUnknownField(t *testing.T) {
	path := filepath.Join("testdata", "unknown_contract.yml")
	_, err := (ContractParser{}).Parse(path)
	if err == nil || !strings.Contains(err.Error(), "field unknown not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestContractParserReturnsReadError(t *testing.T) {
	_, err := (ContractParser{}).Parse(filepath.Join(t.TempDir(), "missing.yml"))
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestParseContractDoesNotValidateRules(t *testing.T) {
	parsed, err := parseContract("contract.yml", []byte("env:\n  TOKEN:\n    type:\n      string:\n        pattern: '['\n"))
	if err != nil {
		t.Fatal(err)
	}
	rule := parsed.Env["TOKEN"]
	if rule.Type.Kind != "" || rule.Type.String.Pattern != nil {
		t.Fatalf("parser prepared a rule: %#v", rule)
	}
}
