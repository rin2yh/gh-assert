package parser

import "testing"

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
