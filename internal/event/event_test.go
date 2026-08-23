package event

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rin2yh/gh-assert/internal/test"
)

func TestName(t *testing.T) {
	t.Setenv("GITHUB_EVENT_NAME", Dispatch)
	if got := Name(); got != Dispatch {
		t.Fatalf("got %q, want %q", got, Dispatch)
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

			inputs, err := Inputs()
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, inputs); diff != "" {
				t.Fatalf("inputs mismatch (-want +got):\n%s", diff)
			}
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

			_, err := Inputs()
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

			_, err := Inputs()
			test.AssertError(t, err)
		})
	}
}

func writeEvent(t *testing.T, payload string) string {
	t.Helper()
	return test.WriteFile(t, t.TempDir(), "event.json", payload)
}
