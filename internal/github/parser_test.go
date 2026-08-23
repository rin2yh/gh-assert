package github

import (
	"path/filepath"
	"testing"
)

func TestParserEventSyntaxes(t *testing.T) {
	tests := []struct {
		name    string
		on      string
		event   string
		present bool
	}{
		{name: "scalar", on: "push", event: "push", present: true},
		{name: "sequence", on: "[push, pull_request]", event: "pull_request", present: true},
		{name: "different scalar", on: "workflow_dispatch", event: "workflow_call", present: false},
		{name: "null mapping value", on: "\n  workflow_call:\n", event: "workflow_call", present: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflow, err := NewParser("workflow.yml").parse([]byte("on: " + tt.on + "\n"))
			if err != nil {
				t.Fatal(err)
			}
			_, present := workflow.Events[tt.event]
			if present != tt.present {
				t.Fatalf("event %q presence = %t, want %t", tt.event, present, tt.present)
			}
		})
	}
}

func TestParserInputs(t *testing.T) {
	path := filepath.Join("testdata", "workflow.yml")
	workflow, err := NewParser(path).Parse()
	if err != nil {
		t.Fatal(err)
	}
	input := workflow.Events["workflow_call"].Inputs["environment"]
	if !input.Required || input.Type != "string" {
		t.Fatalf("unexpected input: %#v", input)
	}
}

func TestParserReturnsReadError(t *testing.T) {
	_, err := NewParser(filepath.Join(t.TempDir(), "missing.yml")).Parse()
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestActionParser(t *testing.T) {
	path := filepath.Join("testdata", "action.yml")
	action, err := NewActionParser(path).Parse()
	if err != nil {
		t.Fatal(err)
	}
	if action.Runs.Using != "composite" {
		t.Fatalf("runs.using = %q, want composite", action.Runs.Using)
	}
	if !action.Inputs["environment"].Required {
		t.Fatal("environment input is not required")
	}
}

func TestActionParserReturnsReadError(t *testing.T) {
	_, err := NewActionParser(filepath.Join(t.TempDir(), "missing.yml")).Parse()
	if err == nil {
		t.Fatal("expected an error")
	}
}
