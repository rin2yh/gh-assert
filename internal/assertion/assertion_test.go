package assertion

import (
	"testing"

	"github.com/rin2yh/gh-assert/internal/contract"
)

func loadContract(t *testing.T, name string) *contract.Contract {
	t.Helper()
	c, err := contract.LoadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func loadRuntimeContract(t *testing.T) *contract.Contract {
	t.Helper()
	return loadContract(t, "runtime.yml")
}

func assertViolationCount(t *testing.T, c *contract.Contract, env values, want int) {
	t.Helper()
	if got := len(Validate(c, env, values(nil))); got != want {
		t.Fatalf("got %d violations, want %d", got, want)
	}
}

func assertInputViolationCount(t *testing.T, c *contract.Contract, inputs values, want int) {
	t.Helper()
	if got := len(Validate(c, values(nil), inputs)); got != want {
		t.Fatalf("got %d violations, want %d", got, want)
	}
}

func TestValidateRequiredEnvironment(t *testing.T) {
	c := loadRuntimeContract(t)
	tests := []struct {
		name string
		env  values
		want int
	}{
		{"required missing", values{"COUNT": "3", "FLAG": "false"}, 1},
		{"required empty", values{"NAME": "", "COUNT": "3", "FLAG": "false"}, 1},
		{"optional missing", values{"NAME": "staging", "FLAG": "true"}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assertViolationCount(t, c, tt.env, tt.want) })
	}
}

func TestValidateStringConstraints(t *testing.T) {
	c := loadRuntimeContract(t)
	tests := []struct {
		name  string
		value string
		want  int
	}{
		{"valid", "staging", 0},
		{"enum and pattern", "production1", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := values{"NAME": tt.value, "COUNT": "3", "FLAG": "false"}
			assertViolationCount(t, c, env, tt.want)
		})
	}
}

func TestValidateIntegerConstraints(t *testing.T) {
	c := loadRuntimeContract(t)
	tests := []struct{ name, value string }{
		{"above maximum", "6"},
		{"invalid type", "three"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertViolationCount(t, c, values{"NAME": "staging", "COUNT": tt.value, "FLAG": "false"}, 1)
		})
	}
}

func TestValidateBooleanType(t *testing.T) {
	c := loadRuntimeContract(t)
	assertViolationCount(t, c, values{"NAME": "staging", "COUNT": "3", "FLAG": "yes"}, 1)
}

func TestValidateWorkflowInputs(t *testing.T) {
	c := loadContract(t, "inputs.yml")
	tests := []struct {
		name   string
		inputs values
		want   int
	}{
		{"valid", values{"environment": "staging", "retries": "3", "dry-run": "true"}, 0},
		{"required missing", values{"retries": "3"}, 1},
		{"required empty", values{"environment": "", "retries": "3"}, 1},
		{"optional missing", values{"environment": "production"}, 0},
		{"enum", values{"environment": "develop"}, 1},
		{"integer above maximum", values{"environment": "staging", "retries": "6"}, 1},
		{"boolean", values{"environment": "staging", "dry-run": "yes"}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assertInputViolationCount(t, c, tt.inputs, tt.want) })
	}
}

func TestValidateReportsScopeAndPosition(t *testing.T) {
	c := loadContract(t, "inputs.yml")
	violations := Validate(c, values(nil), values{"environment": "develop"})
	if len(violations) != 1 {
		t.Fatalf("got %d violations, want 1", len(violations))
	}
	got := violations[0]
	if got.Scope != ScopeInput || got.Name != "environment" {
		t.Fatalf("got scope %q name %q, want %q environment", got.Scope, got.Name, ScopeInput)
	}
	if got.Position.Line == 0 {
		t.Fatal("violation has no position")
	}
}

func TestValidateSeparatesEnvironmentAndInputs(t *testing.T) {
	c := loadContract(t, "combined.yml")
	violations := Validate(c, values{"NAME": "develop"}, values{"environment": "develop"})
	if len(violations) != 2 {
		t.Fatalf("got %d violations, want 2", len(violations))
	}
	if violations[0].Scope != ScopeEnv || violations[1].Scope != ScopeInput {
		t.Fatalf("unexpected scopes: %q, %q", violations[0].Scope, violations[1].Scope)
	}
}

func TestValidateRuntimeReadsProcessEnvironment(t *testing.T) {
	t.Setenv("NAME", "staging")
	t.Setenv("COUNT", "3")
	t.Setenv("FLAG", "true")
	if violations := ValidateRuntime(loadRuntimeContract(t), nil); len(violations) != 0 {
		t.Fatalf("got %d violations, want 0", len(violations))
	}
}
