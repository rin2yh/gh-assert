package github

import (
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
		name    string
		payload string
		want    map[string]string
	}{
		{
			name:    "scalar values",
			payload: `{"inputs":{"environment":"staging","retries":3,"dry-run":true,"note":null}}`,
			want:    map[string]string{"environment": "staging", "retries": "3", "dry-run": "true", "note": ""},
		},
		{
			name:    "no inputs key",
			payload: `{"ref":"refs/heads/main"}`,
			want:    map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GITHUB_EVENT_PATH", writeEvent(t, tt.payload))

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
	tests := []struct{ name, payload string }{
		{name: "invalid json", payload: "{"},
		{name: "non scalar input", payload: `{"inputs":{"matrix":["a"]}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GITHUB_EVENT_PATH", writeEvent(t, tt.payload))

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

func writeEvent(t *testing.T, payload string) string {
	t.Helper()
	return test.WriteFile(t, t.TempDir(), "event.json", payload)
}
