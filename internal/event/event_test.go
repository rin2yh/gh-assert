package event

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
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

func TestInputsRejectsUnreadablePayload(t *testing.T) {
	tests := []struct {
		name string
		path func(t *testing.T) string
	}{
		{name: "no event path", path: func(*testing.T) string { return "" }},
		{name: "missing file", path: func(t *testing.T) string { return filepath.Join(t.TempDir(), "absent.json") }},
		{name: "invalid json", path: func(t *testing.T) string { return writeEvent(t, "{") }},
		{name: "non scalar input", path: func(t *testing.T) string { return writeEvent(t, `{"inputs":{"matrix":["a"]}}`) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GITHUB_EVENT_PATH", tt.path(t))

			if _, err := Inputs(); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func writeEvent(t *testing.T, payload string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "event.json")
	if err := os.WriteFile(path, []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
