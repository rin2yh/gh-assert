package event

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func writeEvent(t *testing.T, payload string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "event.json")
	if err := os.WriteFile(path, []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestName(t *testing.T) {
	t.Setenv("GITHUB_EVENT_NAME", Dispatch)
	if got := Name(); got != Dispatch {
		t.Fatalf("got %q, want %q", got, Dispatch)
	}
}

func TestInputsReadsEventPayload(t *testing.T) {
	t.Setenv("GITHUB_EVENT_PATH", writeEvent(t, `{"inputs":{"environment":"staging","retries":3,"dry-run":true,"note":null}}`))

	inputs, err := Inputs()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"environment": "staging", "retries": "3", "dry-run": "true", "note": ""}
	if diff := cmp.Diff(want, inputs); diff != "" {
		t.Fatalf("inputs mismatch (-want +got):\n%s", diff)
	}
}

func TestInputsWithoutInputsKey(t *testing.T) {
	t.Setenv("GITHUB_EVENT_PATH", writeEvent(t, `{"ref":"refs/heads/main"}`))

	inputs, err := Inputs()
	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) != 0 {
		t.Fatalf("got %d inputs, want 0", len(inputs))
	}
}

func TestInputsRejectsUnreadablePayload(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		noPath  bool
	}{
		{name: "no event path", noPath: true},
		{name: "invalid json", payload: "{"},
		{name: "non scalar input", payload: `{"inputs":{"matrix":["a"]}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.noPath {
				t.Setenv("GITHUB_EVENT_PATH", "")
			} else {
				t.Setenv("GITHUB_EVENT_PATH", writeEvent(t, tt.payload))
			}
			if _, err := Inputs(); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestInputsRejectsMissingFile(t *testing.T) {
	t.Setenv("GITHUB_EVENT_PATH", filepath.Join(t.TempDir(), "absent.json"))

	if _, err := Inputs(); err == nil {
		t.Fatal("expected an error")
	}
}
