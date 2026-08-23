package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/rin2yh/gh-assert/internal/test"
)

func TestValidateContract(t *testing.T) {
	path := writeContract(t, "env: {}\n")
	stdout, _ := runCommand(t, []string{"validate", path}, 0)

	assertContains(t, stdout, "contract is valid", true)
}

func TestValidateReportsOutputError(t *testing.T) {
	path := writeContract(t, "env: {}\n")
	var stderr strings.Builder
	code := execute([]string{"validate", path}, errorWriter{}, &stderr)

	assertExitCode(t, code, 2, stderr.String())
}

func TestValidateRejectsInvalidContract(t *testing.T) {
	path := writeContract(t, "env: []\n")
	_, stderr := runCommand(t, []string{"validate", path}, 2)

	assertContains(t, stderr, "cannot unmarshal", true)
}

func TestValidateReusableWorkflowInterface(t *testing.T) {
	stdout, _ := runCommand(t, []string{"validate", "testdata/reusable/deploy_assert.yml"}, 0)

	assertContains(t, stdout, "contract is valid", true)
}

func TestValidateRejectsReusableWorkflowInterfaceMismatch(t *testing.T) {
	_, stderr := runCommand(t, []string{"validate", "testdata/reusable/mismatch_assert.yml"}, 2)

	assertContains(t, stderr, "reusable workflow interface does not match contract", true)
}

func TestRuntimeSuccess(t *testing.T) {
	t.Setenv("FLAG", "true")
	path := writeContract(t, "env:\n  FLAG:\n    required: true\n    type:\n      boolean: {}\n")

	runCommand(t, []string{"--contract", path}, 0)
}

func TestRuntimeFailureHidesValue(t *testing.T) {
	const secret = "secret-value"
	t.Setenv("TOKEN", secret)
	path := writeContract(t, "env:\n  TOKEN:\n    required: true\n    type:\n      string:\n        enum: [expected]\n")
	_, stderr := runCommand(t, []string{"--contract", path}, 1)

	assertContains(t, stderr, "allowed enum", true)
	assertContains(t, stderr, secret, false)
}

func TestRuntimeReportsDiagnosticOutputError(t *testing.T) {
	t.Setenv("TOKEN", "unexpected")
	path := writeContract(t, "env:\n  TOKEN:\n    required: true\n    type:\n      string:\n        enum: [expected]\n")
	code := execute([]string{"--contract", path}, &strings.Builder{}, errorWriter{})

	assertExitCode(t, code, 2, "")
}

func TestRuntimeAssertsWorkflowDispatchInputs(t *testing.T) {
	setEvent(t, `{"inputs":{"environment":"develop"}}`)
	path := writeContract(t, "inputs:\n  environment:\n    required: true\n    type:\n      string:\n        enum: [staging, production]\n")
	_, stderr := runCommand(t, []string{"--contract", path}, 1)

	assertContains(t, stderr, "input environment", true)
	assertContains(t, stderr, "develop", false)
}

func TestRuntimeSucceedsWithValidInputs(t *testing.T) {
	setEvent(t, `{"inputs":{"environment":"staging","retries":3}}`)
	path := writeContract(t, "inputs:\n  environment:\n    required: true\n    type:\n      string:\n        enum: [staging, production]\n  retries:\n    type:\n      integer:\n        min: 0\n        max: 5\n")

	runCommand(t, []string{"--contract", path}, 0)
}

func TestRuntimeAssertsReusableWorkflowInputValues(t *testing.T) {
	t.Setenv("GITHUB_EVENT_NAME", "push")
	t.Setenv("GITHUB_EVENT_PATH", "")
	t.Setenv("FLAG", "true")
	t.Setenv("GH_ASSERT_INPUTS", `{"environment":"develop"}`)

	_, stderr := runCommand(t, []string{"--contract", "testdata/reusable/deploy_assert.yml"}, 1)

	assertContains(t, stderr, "allowed enum", true)
	assertContains(t, stderr, "develop", false)
}

func TestRuntimeReusableWorkflowRequiresInputValues(t *testing.T) {
	t.Setenv("GH_ASSERT_INPUTS", "")

	_, stderr := runCommand(t, []string{"--contract", "testdata/reusable/deploy_assert.yml"}, 2)

	assertContains(t, stderr, "workflow-inputs", true)
}

func TestRuntimeRejectsInputsContractOutsideWorkflowDispatch(t *testing.T) {
	tests := []struct {
		name      string
		eventName string
	}{
		{name: "other event", eventName: "push"},
		{name: "no event name", eventName: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GITHUB_EVENT_NAME", tt.eventName)
			t.Setenv("GITHUB_EVENT_PATH", "")
			path := writeContract(t, "inputs:\n  environment:\n    required: true\n    type:\n      string: {}\n")
			_, stderr := runCommand(t, []string{"--contract", path}, 2)

			assertContains(t, stderr, "workflow_dispatch", true)
		})
	}
}

func TestRuntimeIgnoresEventWithoutInputRules(t *testing.T) {
	tests := []struct{ name, contract string }{
		{name: "empty inputs section", contract: "inputs: {}\n"},
		{name: "env only", contract: "env:\n  FLAG:\n    required: true\n    type:\n      boolean: {}\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GITHUB_EVENT_NAME", "push")
			t.Setenv("GITHUB_EVENT_PATH", "")
			t.Setenv("FLAG", "true")

			runCommand(t, []string{"--contract", writeContract(t, tt.contract)}, 0)
		})
	}
}

func TestValidateAcceptsInputsContract(t *testing.T) {
	path := writeContract(t, "inputs:\n  environment:\n    required: true\n    type:\n      string:\n        enum: [staging, production]\n")
	stdout, _ := runCommand(t, []string{"validate", path}, 0)

	assertContains(t, stdout, "contract is valid", true)
}

func TestHelp(t *testing.T) {
	stdout, _ := runCommand(t, []string{"help"}, 0)

	assertContains(t, stdout, "validate", true)
}

func TestRejectsUnknownCommand(t *testing.T) {
	_, stderr := runCommand(t, []string{"unknown"}, 2)

	assertContains(t, stderr, "unknown", true)
}

func TestValidateRejectsExtraArgument(t *testing.T) {
	_, stderr := runCommand(t, []string{"validate", "one.yml", "two.yml"}, 2)

	assertContains(t, stderr, "at most 1 arg", true)
}

func runCommand(t *testing.T, args []string, wantCode int) (string, string) {
	t.Helper()
	var stdout, stderr strings.Builder
	code := execute(args, &stdout, &stderr)
	assertExitCode(t, code, wantCode, stderr.String())
	return stdout.String(), stderr.String()
}

func assertExitCode(t *testing.T, got, want int, stderr string) {
	t.Helper()
	if got != want {
		t.Fatalf("exit code = %d, want %d; stderr=%q", got, want, stderr)
	}
}

func assertContains(t *testing.T, got, want string, wantContains bool) {
	t.Helper()
	if contains := strings.Contains(got, want); contains != wantContains {
		t.Errorf("output = %q, contains %q = %v, want %v", got, want, contains, wantContains)
	}
}

func setEvent(t *testing.T, payload string) {
	t.Helper()
	t.Setenv("GITHUB_EVENT_NAME", "workflow_dispatch")
	t.Setenv("GITHUB_EVENT_PATH", test.WriteFile(t, t.TempDir(), "event.json", payload))
}

func writeContract(t *testing.T, content string) string {
	t.Helper()
	return test.WriteFile(t, t.TempDir(), "contract.yml", content)
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }
