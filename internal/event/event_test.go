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

func TestDispatchInputsReadsEventPayload(t *testing.T) {
	path := writeEvent(t, `{"inputs":{"environment":"staging","retries":3,"dry-run":true,"note":null}}`)
	t.Setenv("GITHUB_EVENT_NAME", "workflow_dispatch")
	t.Setenv("GITHUB_EVENT_PATH", path)

	inputs, err := DispatchInputs()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"environment": "staging", "retries": "3", "dry-run": "true", "note": ""}
	if diff := cmp.Diff(want, inputs); diff != "" {
		t.Fatalf("inputs mismatch (-want +got):\n%s", diff)
	}
}

func TestDispatchInputsWithoutInputsKey(t *testing.T) {
	t.Setenv("GITHUB_EVENT_NAME", "workflow_dispatch")
	t.Setenv("GITHUB_EVENT_PATH", writeEvent(t, `{"ref":"refs/heads/main"}`))

	inputs, err := DispatchInputs()
	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) != 0 {
		t.Fatalf("got %d inputs, want 0", len(inputs))
	}
}

func TestDispatchInputsRejectsUnavailableEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventName string
		payload   string
		noPath    bool
	}{
		{name: "no event name", eventName: ""},
		{name: "other event", eventName: "push"},
		{name: "no event path", eventName: "workflow_dispatch", noPath: true},
		{name: "invalid json", eventName: "workflow_dispatch", payload: "{"},
		{name: "non scalar input", eventName: "workflow_dispatch", payload: `{"inputs":{"matrix":["a"]}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GITHUB_EVENT_NAME", tt.eventName)
			if tt.noPath {
				t.Setenv("GITHUB_EVENT_PATH", "")
			} else {
				t.Setenv("GITHUB_EVENT_PATH", writeEvent(t, tt.payload))
			}
			if _, err := DispatchInputs(); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestDispatchInputsRejectsMissingFile(t *testing.T) {
	t.Setenv("GITHUB_EVENT_NAME", "workflow_dispatch")
	t.Setenv("GITHUB_EVENT_PATH", filepath.Join(t.TempDir(), "absent.json"))

	if _, err := DispatchInputs(); err == nil {
		t.Fatal("expected an error")
	}
}
