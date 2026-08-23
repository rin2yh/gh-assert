package assertion

import (
	"testing"

	"github.com/rin2yh/gh-assert/internal/contract"
	"github.com/rin2yh/gh-assert/internal/model"
)

func TestValidateRequiredEnvironment(t *testing.T) {
	c := loadContract(t, "runtime.yml")
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
		t.Run(tt.name, func(t *testing.T) {
			assertViolationCount(t, Validate(c, tt.env, nil), tt.want)
		})
	}
}

func TestValidateEnvironmentConstraints(t *testing.T) {
	c := loadContract(t, "runtime.yml")
	tests := []struct {
		name string
		env  values
		want int
	}{
		{"valid", values{"NAME": "staging", "COUNT": "3", "FLAG": "false"}, 0},
		{"enum and pattern", values{"NAME": "production1", "COUNT": "3", "FLAG": "false"}, 2},
		{"integer above maximum", values{"NAME": "staging", "COUNT": "6", "FLAG": "false"}, 1},
		{"integer invalid type", values{"NAME": "staging", "COUNT": "three", "FLAG": "false"}, 1},
		{"boolean", values{"NAME": "staging", "COUNT": "3", "FLAG": "yes"}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertViolationCount(t, Validate(c, tt.env, nil), tt.want)
		})
	}
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
		t.Run(tt.name, func(t *testing.T) {
			assertViolationCount(t, Validate(c, nil, tt.inputs), tt.want)
		})
	}
}

func TestValidateReportsScopeAndPosition(t *testing.T) {
	c := loadContract(t, "inputs.yml")
	violations := Validate(c, values(nil), values{"environment": "develop"})

	assertScopes(t, violations, ScopeInput)
	if violations[0].Name != "environment" {
		t.Errorf("violation name = %q, want environment", violations[0].Name)
	}
	if violations[0].Position.Line == 0 {
		t.Error("violation has no position")
	}
}

func TestValidateSeparatesEnvironmentAndInputs(t *testing.T) {
	c := loadContract(t, "combined.yml")

	assertScopes(t, Validate(c, values{"NAME": "develop"}, values{"environment": "develop"}), ScopeEnv, ScopeInput)
}

func TestValidateRuntimeReadsProcessEnvironment(t *testing.T) {
	t.Setenv("NAME", "staging")
	t.Setenv("COUNT", "3")
	t.Setenv("FLAG", "true")

	assertViolationCount(t, ValidateRuntime(loadContract(t, "runtime.yml"), nil), 0)
}

func loadContract(t *testing.T, name string) *model.Contract {
	t.Helper()
	c, err := contract.LoadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func assertViolationCount(t *testing.T, violations []Violation, want int) {
	t.Helper()
	if len(violations) != want {
		t.Fatalf("got %d violations, want %d: %v", len(violations), want, violations)
	}
}

func assertScopes(t *testing.T, violations []Violation, want ...string) {
	t.Helper()
	assertViolationCount(t, violations, len(want))
	for i, scope := range want {
		if violations[i].Scope != scope {
			t.Errorf("violation %d scope = %q, want %q", i, violations[i].Scope, scope)
		}
	}
}
