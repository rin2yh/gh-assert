package github

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rin2yh/gh-assert/internal/test"
)

func TestName(t *testing.T) {
	t.Setenv("GITHUB_EVENT_NAME", WorkflowDispatch)
	if got := EventName(); got != WorkflowDispatch {
		t.Fatalf("got %q, want %q", got, WorkflowDispatch)
	}
}

func TestInputsReadsEventPayload(t *testing.T) {
	tests := []struct {
		name string
		file string
		want map[string]string
	}{
		{
			name: "scalar values",
			file: "scalar-inputs.json",
			want: map[string]string{"environment": "staging", "retries": "3", "dry-run": "true", "note": ""},
		},
		{
			name: "no inputs key",
			file: "no-inputs.json",
			want: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GITHUB_EVENT_PATH", filepath.Join("testdata", tt.file))

			inputs, err := EventInputs()
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, inputs); diff != "" {
				t.Fatalf("inputs mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWorkflowInputsReadsInputsContext(t *testing.T) {
	t.Setenv(InputsJSON, `{"environment":"staging","retries":3,"dry-run":true,"note":null}`)

	inputs, err := WorkflowInputs()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"environment": "staging", "retries": "3", "dry-run": "true", "note": ""}
	if diff := cmp.Diff(want, inputs); diff != "" {
		t.Fatalf("inputs mismatch (-want +got):\n%s", diff)
	}
}

func TestWorkflowInputsRejectsInvalidContext(t *testing.T) {
	tests := []struct{ name, inputs string }{
		{name: "invalid json", inputs: "{"},
		{name: "not an object", inputs: `[]`},
		{name: "non scalar input", inputs: `{"matrix":["a"]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(InputsJSON, tt.inputs)

			_, err := WorkflowInputs()
			test.AssertError(t, err)
		})
	}
}

func TestInputsRejectsInvalidPayload(t *testing.T) {
	tests := []struct{ name, file string }{
		{name: "invalid json", file: "invalid.json"},
		{name: "non scalar input", file: "non-scalar-input.json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GITHUB_EVENT_PATH", filepath.Join("testdata", tt.file))

			_, err := EventInputs()
			test.AssertError(t, err)
		})
	}
}

func TestInputsRejectsUnavailablePayload(t *testing.T) {
	tests := []struct{ name, path string }{
		{name: "no event path", path: ""},
		{name: "missing file", path: "testdata/absent.json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GITHUB_EVENT_PATH", tt.path)

			_, err := EventInputs()
			test.AssertError(t, err)
		})
	}
}
