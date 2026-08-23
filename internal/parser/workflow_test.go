package parser

import "testing"

func TestParseWorkflowEventSyntaxes(t *testing.T) {
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
			workflow, err := parseWorkflow("workflow.yml", []byte("on: "+tt.on+"\n"))
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

func TestParseWorkflowInputs(t *testing.T) {
	workflow, err := parseWorkflow("workflow.yml", []byte("on:\n  workflow_call:\n    inputs:\n      environment:\n        required: true\n        type: string\n"))
	if err != nil {
		t.Fatal(err)
	}
	input := workflow.Events["workflow_call"].Inputs["environment"]
	if !input.Required || input.Type != "string" {
		t.Fatalf("unexpected input: %#v", input)
	}
}
