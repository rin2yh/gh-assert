package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateContract(t *testing.T) {
	path := writeContract(t, "env: {}\n")
	code, stdout, stderr := runCommand([]string{"validate", path})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "contract is valid") {
		t.Errorf("stdout = %q, want contract success message", stdout)
	}
}

func TestValidateReportsOutputError(t *testing.T) {
	path := writeContract(t, "env: {}\n")
	code := execute([]string{"validate", path}, errorWriter{}, &strings.Builder{})

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
}

func TestValidateRejectsInvalidContract(t *testing.T) {
	path := writeContract(t, "env: []\n")
	code, _, stderr := runCommand([]string{"validate", path})

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr, "cannot unmarshal") {
		t.Errorf("stderr = %q, want YAML error", stderr)
	}
}

func TestRuntimeSuccess(t *testing.T) {
	t.Setenv("FLAG", "true")
	path := writeContract(t, "env:\n  FLAG:\n    required: true\n    type:\n      boolean: {}\n")
	code, _, stderr := runCommand([]string{"--contract", path})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
}

func TestRuntimeFailureHidesValue(t *testing.T) {
	const secret = "secret-value"
	t.Setenv("TOKEN", secret)
	path := writeContract(t, "env:\n  TOKEN:\n    required: true\n    type:\n      string:\n        enum: [expected]\n")
	code, _, stderr := runCommand([]string{"--contract", path})

	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr, "allowed enum") {
		t.Errorf("stderr = %q, want enum violation", stderr)
	}
	if strings.Contains(stderr, secret) {
		t.Errorf("stderr contains hidden value %q", secret)
	}
}

func TestRuntimeReportsDiagnosticOutputError(t *testing.T) {
	t.Setenv("TOKEN", "unexpected")
	path := writeContract(t, "env:\n  TOKEN:\n    required: true\n    type:\n      string:\n        enum: [expected]\n")
	code := execute([]string{"--contract", path}, &strings.Builder{}, errorWriter{})

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
}

func TestRuntimeAssertsWorkflowDispatchInputs(t *testing.T) {
	setEvent(t, `{"inputs":{"environment":"develop"}}`)
	path := writeContract(t, "inputs:\n  environment:\n    required: true\n    type:\n      string:\n        enum: [staging, production]\n")
	code, _, stderr := runCommand([]string{"--contract", path})

	if code != 1 {
		t.Fatalf("exit code = %d, want 1; stderr=%q", code, stderr)
	}
	if !strings.Contains(stderr, "input environment") {
		t.Errorf("stderr = %q, want an input violation", stderr)
	}
	if strings.Contains(stderr, "develop") {
		t.Errorf("stderr contains the input value: %q", stderr)
	}
}

func TestRuntimeSucceedsWithValidInputs(t *testing.T) {
	setEvent(t, `{"inputs":{"environment":"staging","retries":3}}`)
	path := writeContract(t, "inputs:\n  environment:\n    required: true\n    type:\n      string:\n        enum: [staging, production]\n  retries:\n    type:\n      integer:\n        min: 0\n        max: 5\n")
	code, _, stderr := runCommand([]string{"--contract", path})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
}

func TestRuntimeRejectsInputsOutsideWorkflowDispatch(t *testing.T) {
	t.Setenv("GITHUB_EVENT_NAME", "push")
	t.Setenv("GITHUB_EVENT_PATH", "")
	path := writeContract(t, "inputs:\n  environment:\n    required: true\n    type:\n      string: {}\n")
	code, _, stderr := runCommand([]string{"--contract", path})

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr, "workflow_dispatch") {
		t.Errorf("stderr = %q, want a workflow_dispatch requirement error", stderr)
	}
}

func TestRuntimeAllowsEmptyInputsSectionOutsideWorkflowDispatch(t *testing.T) {
	t.Setenv("GITHUB_EVENT_NAME", "push")
	t.Setenv("GITHUB_EVENT_PATH", "")
	path := writeContract(t, "inputs: {}\n")
	code, _, stderr := runCommand([]string{"--contract", path})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
}

func TestRuntimeIgnoresEventWithoutInputsContract(t *testing.T) {
	t.Setenv("GITHUB_EVENT_NAME", "push")
	t.Setenv("FLAG", "true")
	path := writeContract(t, "env:\n  FLAG:\n    required: true\n    type:\n      boolean: {}\n")
	code, _, stderr := runCommand([]string{"--contract", path})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
}

func TestValidateAcceptsInputsContract(t *testing.T) {
	path := writeContract(t, "inputs:\n  environment:\n    required: true\n    type:\n      string:\n        enum: [staging, production]\n")
	code, stdout, stderr := runCommand([]string{"validate", path})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "contract is valid") {
		t.Errorf("stdout = %q, want contract success message", stdout)
	}
}

func TestHelp(t *testing.T) {
	code, stdout, stderr := runCommand([]string{"help"})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "validate") {
		t.Errorf("stdout = %q, want validate command", stdout)
	}
}

func TestRejectsUnknownCommand(t *testing.T) {
	code, _, stderr := runCommand([]string{"unknown"})

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr, "unknown") {
		t.Errorf("stderr = %q, want unknown command error", stderr)
	}
}

func TestValidateRejectsExtraArgument(t *testing.T) {
	code, _, stderr := runCommand([]string{"validate", "one.yml", "two.yml"})

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr, "at most 1 arg") {
		t.Errorf("stderr = %q, want argument count error", stderr)
	}
}

func setEvent(t *testing.T, payload string) {
	t.Helper()
	t.Setenv("GITHUB_EVENT_NAME", "workflow_dispatch")
	t.Setenv("GITHUB_EVENT_PATH", writeFile(t, "event.json", payload))
}

func writeContract(t *testing.T, content string) string {
	t.Helper()
	return writeFile(t, "contract.yml", content)
}

func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func runCommand(args []string) (int, string, string) {
	var stdout, stderr strings.Builder
	code := execute(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }
