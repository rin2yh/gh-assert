package runtime

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rin2yh/gh-assert/internal/test"
)

func TestIsReusableWorkflow(t *testing.T) {
	tests := []struct {
		name     string
		workflow string
		want     bool
	}{
		{name: "workflow call", workflow: "on:\n  workflow_call:\njobs:\n  test:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo test\n", want: true},
		{name: "workflow dispatch", workflow: "on: workflow_dispatch\njobs:\n  test:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo test\n", want: false},
		{name: "missing workflow", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := filepath.Join(t.TempDir(), ".github", "workflows")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			if tt.workflow != "" {
				test.WriteFile(t, dir, "deploy.yml", tt.workflow)
			}

			reusable, err := isReusableWorkflow(filepath.Join(dir, "deploy.yml"))
			if err != nil {
				t.Fatal(err)
			}
			if reusable != tt.want {
				t.Errorf("isReusableWorkflow() = %v, want %v", reusable, tt.want)
			}
		})
	}
}

func TestLoadEventInputsReadsEventPayload(t *testing.T) {
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

			inputs, err := loadEventInputs()
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, inputs); diff != "" {
				t.Fatalf("inputs mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLoadWorkflowInputsReadsInputsContext(t *testing.T) {
	t.Setenv(inputsJSON, `{"environment":"staging","retries":3,"dry-run":true,"note":null}`)

	inputs, err := loadForwardedInputs()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"environment": "staging", "retries": "3", "dry-run": "true", "note": ""}
	if diff := cmp.Diff(want, inputs); diff != "" {
		t.Fatalf("inputs mismatch (-want +got):\n%s", diff)
	}
}

func TestLoadWorkflowInputsRejectsInvalidContext(t *testing.T) {
	tests := []struct{ name, inputs string }{
		{name: "invalid json", inputs: "{"},
		{name: "not an object", inputs: `[]`},
		{name: "non scalar input", inputs: `{"matrix":["a"]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(inputsJSON, tt.inputs)

			_, err := loadForwardedInputs()
			test.AssertError(t, err)
		})
	}
}

func TestLoadEventInputsRejectsInvalidPayload(t *testing.T) {
	tests := []struct{ name, file string }{
		{name: "invalid json", file: "invalid.json"},
		{name: "non scalar input", file: "non-scalar-input.json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GITHUB_EVENT_PATH", filepath.Join("testdata", tt.file))

			_, err := loadEventInputs()
			test.AssertError(t, err)
		})
	}
}

func TestLoadEventInputsRejectsUnavailablePayload(t *testing.T) {
	tests := []struct{ name, path string }{
		{name: "no event path", path: ""},
		{name: "missing file", path: "testdata/absent.json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GITHUB_EVENT_PATH", tt.path)

			_, err := loadEventInputs()
			test.AssertError(t, err)
		})
	}
}
